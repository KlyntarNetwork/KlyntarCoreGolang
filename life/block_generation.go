package life

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/system_contracts"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
	"github.com/gorilla/websocket"
)

type DoubleMap = map[string]map[string]string

var LEADER_ROTATION_PROOFS DoubleMap

var WEBSOCKET_CONNECTIONS_FOR_ALRP map[string]*websocket.Conn

type RotationProofCollector struct {
	wsConnMap map[string]*websocket.Conn
	quorum    []string
	majority  int
	timeout   time.Duration
}

// To grab proofs for multiple previous leaders in a parallel way
func (c *RotationProofCollector) AlrpForLeadersCollector(ctx context.Context, leaderIDs []string, messageBuilder func(leaderID string) []byte) DoubleMap {

	var wg sync.WaitGroup
	mu := sync.Mutex{}

	result := make(DoubleMap)

	for _, leaderID := range leaderIDs {
		wg.Add(1)

		go func(leaderID string) {
			defer wg.Done()

			waiter := utils.NewQuorumWaiter(len(c.quorum))

			// Create a timeout for a call
			leaderCtx, cancel := context.WithTimeout(ctx, c.timeout)
			defer cancel()

			message := messageBuilder(leaderID)

			responses, ok := waiter.SendAndWait(leaderCtx, message, c.quorum, c.wsConnMap, c.majority)
			if !ok {
				return
			}

			// Build final result
			mapping := make(map[string]string, len(responses))
			for validatorID, raw := range responses {
				mapping[validatorID] = string(raw)
			}

			mu.Lock()
			result[leaderID] = mapping
			mu.Unlock()
		}(leaderID)
	}

	wg.Wait()
	return result
}

func BlocksGenerationThread() {

	generateBlocksPortion()

	time.AfterFunc(time.Duration(globals.APPROVEMENT_THREAD.Thread.NetworkParameters.BlockTime), func() {
		BlocksGenerationThread()
	})

}

func getTransactionsFromMempool() []structures.Transaction {

	globals.MEMPOOL.Mutex.Lock()
	defer globals.MEMPOOL.Mutex.Unlock()

	globals.APPROVEMENT_THREAD.RWMutex.RLock()
	defer globals.APPROVEMENT_THREAD.RWMutex.RUnlock()

	limit := globals.APPROVEMENT_THREAD.Thread.NetworkParameters.TxLimitPerBlock

	if limit > len(globals.MEMPOOL.Slice) {
		limit = len(globals.MEMPOOL.Slice)
	}

	transactions := make([]structures.Transaction, limit)

	copy(transactions, globals.MEMPOOL.Slice[:limit])

	globals.MEMPOOL.Slice = globals.MEMPOOL.Slice[limit:]

	return transactions
}

func getAggregatedEpochFinalizationProof(epochHandler *structures.EpochHandler) *structures.AggregatedEpochFinalizationProof {

	previousEpochIndex := epochHandler.Id - 1

	// Try to find locally

	aefpProofRaw, err := globals.EPOCH_DATA.Get([]byte("AEFP:"+strconv.Itoa(previousEpochIndex)), nil)

	aefpParsed := new(structures.AggregatedEpochFinalizationProof)

	if parsErr := json.Unmarshal(aefpProofRaw, aefpParsed); parsErr == nil && err == nil {

		return aefpParsed

	}

	quorumUrlsAndPubkeys := common_functions.GetQuorumUrlsAndPubkeys(epochHandler)

	var quorumUrls []string

	for _, quorumMember := range quorumUrlsAndPubkeys {

		quorumUrls = append(quorumUrls, quorumMember.Url)

	}

	allKnownNodes := append(quorumUrls, globals.CONFIGURATION.BootstrapNodes...)

	legacyEpochHandlerRaw, err := globals.EPOCH_DATA.Get([]byte("EPOCH_HANDLER:"+strconv.Itoa(previousEpochIndex)), nil)

	if err != nil {
		return nil
	}

	legacyEpochHandler := new(structures.EpochHandler)

	errParse := json.Unmarshal(legacyEpochHandlerRaw, legacyEpochHandler)

	if errParse != nil {
		return nil
	}

	legacyEpochFullID := legacyEpochHandler.Hash + "#" + strconv.Itoa(legacyEpochHandler.Id)

	legacyMajority := common_functions.GetQuorumMajority(legacyEpochHandler)

	legacyQuorum := legacyEpochHandler.Quorum

	// Prepare requests
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan structures.AggregatedEpochFinalizationProof, 1)

	var wg sync.WaitGroup

	for _, nodeEndpoint := range allKnownNodes {

		wg.Add(1)

		go func(endpoint string) {
			defer wg.Done()

			reqCtx, reqCancel := context.WithTimeout(ctx, 2*time.Second)
			defer reqCancel()

			finalURL := endpoint + "/aggregated_epoch_finalization_proof/" + strconv.Itoa(previousEpochIndex)

			req, err := http.NewRequestWithContext(reqCtx, "GET", finalURL, nil)
			if err != nil {
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return
			}

			var proofCandidate structures.AggregatedEpochFinalizationProof

			if err := json.NewDecoder(resp.Body).Decode(&proofCandidate); err != nil {
				return
			}

			if common_functions.VerifyAggregatedEpochFinalizationProof(&proofCandidate, legacyQuorum, legacyMajority, legacyEpochFullID) {
				select {
				case resultChan <- proofCandidate:
					cancel() // stop other goroutines
				default:
				}
			}
		}(nodeEndpoint)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// We need only first valid result

	aefp, ok := <-resultChan

	if ok {
		return &aefp
	}

	return nil
}

func getAggregatedLeaderRotationProof() *structures.AggregatedLeaderRotationProof {

	return nil

}

func getBatchOfApprovedDelayedTxsByQuorum(indexOfLeader int) system_contracts.DelayedTransactionsBatch {

	epochHandler := globals.APPROVEMENT_THREAD.Thread.EpochHandler

	prevEpochIndex := epochHandler.Id - 2

	if indexOfLeader != 0 {

		return system_contracts.DelayedTransactionsBatch{
			EpochIndex:          prevEpochIndex,
			DelayedTransactions: []map[string]string{},
			Proofs:              map[string]string{},
		}

	}

	// var delayedTransactions []map[string]string

	return system_contracts.DelayedTransactionsBatch{}

}

func generateBlocksPortion() {

	globals.APPROVEMENT_THREAD.RWMutex.RLock()

	defer globals.APPROVEMENT_THREAD.RWMutex.RUnlock()

	epochHandler := globals.APPROVEMENT_THREAD.Thread.EpochHandler

	epochFullID := epochHandler.Hash + "#" + strconv.Itoa(epochHandler.Id)

	epochIndex := epochHandler.Id

	currentLeaderPubKey := epochHandler.LeadersSequence[epochHandler.CurrentLeaderIndex]

	/*
			let proofsGrabber = GLOBAL_CACHES.TEMP_CACHE.get(epochIndex+':PROOFS_GRABBER')

		    if(proofsGrabber && WORKING_THREADS.GENERATION_THREAD.epochFullId === epochFullID && WORKING_THREADS.GENERATION_THREAD.nextIndex > proofsGrabber.acceptedIndex+1) return

	*/

	// Safe "if" branch to prevent unnecessary blocks generation

	if currentLeaderPubKey == globals.CONFIGURATION.PublicKey {

		var aefpForPreviousEpoch *structures.AggregatedEpochFinalizationProof = nil

		// Check if <epochFullID> is the same in APPROVEMENT_THREAD and in GENERATION_THREAD

		if globals.GENERATION_THREAD.EpochFullId != epochFullID {

			// If new epoch - add the aggregated proof of previous epoch finalization

			if epochIndex != 0 {

				aefpForPreviousEpoch = getAggregatedEpochFinalizationProof(&epochHandler)

				if aefpForPreviousEpoch == nil {

					return

				}

			}

			// Update the index & hash of epoch

			globals.GENERATION_THREAD.EpochFullId = epochFullID

			// Nullish the index & hash in generation thread for new epoch

			globals.GENERATION_THREAD.PrevHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

			globals.GENERATION_THREAD.NextIndex = 0

		}

	}

}
