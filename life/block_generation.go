package life

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/block"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/system_contracts"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
	"github.com/gorilla/websocket"
	"github.com/syndtr/goleveldb/leveldb"
)

type DoubleMap = map[string]map[string]string

var LEADER_ROTATION_PROOFS DoubleMap // leaderPubkey => map(quorumMemberPubkey=>leaderRotationProofSigna)

var WEBSOCKET_CONNECTIONS_FOR_ALRP map[string]*websocket.Conn // quorumMember => websocket handler

var QUORUM_WAITER *utils.QuorumWaiter

type RotationProofCollector struct {
	wsConnMap map[string]*websocket.Conn
	quorum    []string
	majority  int
	timeout   time.Duration
}

func BlocksGenerationThread() {

	generateBlocksPortion()

	time.AfterFunc(time.Duration(globals.APPROVEMENT_THREAD.Thread.NetworkParameters.BlockTime), func() {
		BlocksGenerationThread()
	})

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

func getTransactionsFromMempool() []structures.Transaction {

	globals.MEMPOOL.Mutex.Lock()
	defer globals.MEMPOOL.Mutex.Unlock()

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
		TODO:

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

		extraData := make(map[string]any)

		if globals.GENERATION_THREAD.NextIndex == 0 {

			if epochIndex > 0 {

				if aefpForPreviousEpoch != nil {

					extraData["aefpForPreviousEpoch"] = aefpForPreviousEpoch

				} else {

					return

				}

			}

			// Build the template to insert to the extraData of block. Structure is {pool0:ALRP,...,poolN:ALRP}

			myIndexInLeadersSequence := slices.Index(epochHandler.LeadersSequence, globals.CONFIGURATION.PublicKey)

			// Get all previous pools - from zero to <my_position>

			pubKeysOfAllThePreviousPools := slices.Clone(epochHandler.LeadersSequence[:myIndexInLeadersSequence])

			// Reverse the slice
			for i, j := 0, len(pubKeysOfAllThePreviousPools)-1; i < j; i, j = i+1, j-1 {
				pubKeysOfAllThePreviousPools[i], pubKeysOfAllThePreviousPools[j] = pubKeysOfAllThePreviousPools[j], pubKeysOfAllThePreviousPools[i]
			}

			indexOfPreviousLeaderToMe := myIndexInLeadersSequence - 1

			previousToMeLeaderPubKey := epochHandler.LeadersSequence[indexOfPreviousLeaderToMe]

			extraData["delayedTxsBatch"] = getBatchOfApprovedDelayedTxsByQuorum(epochHandler.CurrentLeaderIndex)

			//_____________________ Fill the extraData.aggregatedLeadersRotationProofs _____________________

			extraData["aggregatedLeadersRotationProofs"] = make(map[string]structures.AggregatedLeaderRotationProof)

			/*

			   Here we need to fill the object with aggregated leader rotation proofs (ALRPs) for all the previous pools till the pool which was rotated on not-zero height

			   If we can't find all the required ALRPs - skip this iteration to try again later

			*/

			// Add the ALRP for the previous pools in leaders sequence

			pubkeysOfLeadersToGetAlrps := []string{}

			for _, leaderPubKey := range pubKeysOfAllThePreviousPools {

				votingFinalizationStatsPerPool := &structures.PoolVotingStat{
					Index: -1,
				}

				keyBytes := []byte(strconv.Itoa(epochIndex) + ":" + leaderPubKey)

				if finStatsRaw, err := globals.FINALIZATION_VOTING_STATS.Get(keyBytes, nil); err == nil {

					if jsonErrParse := json.Unmarshal(finStatsRaw, votingFinalizationStatsPerPool); jsonErrParse == nil {

						proofThatAtLeastFirstBlockWasCreated := votingFinalizationStatsPerPool.Index >= 0

						// We 100% need ALRP for previous pool
						// But no need in pools who created at least one block in epoch and it's not our previous pool

						if leaderPubKey != previousToMeLeaderPubKey && proofThatAtLeastFirstBlockWasCreated {

							break

						}

					}

				}

				pubkeysOfLeadersToGetAlrps = append(pubkeysOfLeadersToGetAlrps, leaderPubKey)

			}

			// Now when we have a list of previous leader to get ALRP for them - run it

		}

		extraData["rest"] = globals.CONFIGURATION.ExtraDataToBlock

		blockDbAtomicBatch := new(leveldb.Batch)

		blockCandidate := block.NewBlock(getTransactionsFromMempool(), extraData, epochFullID)

		blockHash := blockCandidate.GetHash()

		blockCandidate.Sig = ed25519.GenerateSignature(globals.CONFIGURATION.PrivateKey, blockHash)

		// BlockID has the following format => epochID(epochIndex):Ed25519_Pubkey:IndexOfBlockInCurrentEpoch
		blockID := strconv.Itoa(epochIndex) + ":" + globals.CONFIGURATION.PublicKey + ":" + strconv.Itoa(blockCandidate.Index)

		utils.LogWithTime("New block generated "+blockID, utils.CYAN_COLOR)

		if blockBytes, serializeErr := json.Marshal(blockCandidate); serializeErr == nil {

			globals.GENERATION_THREAD.PrevHash = blockHash

			globals.GENERATION_THREAD.NextIndex++

			if gtBytes, serializeErr2 := json.Marshal(globals.GENERATION_THREAD); serializeErr2 == nil {

				// Store block locally
				blockDbAtomicBatch.Put([]byte(blockID), blockBytes)

				// Update the GENERATION_THREAD after all
				blockDbAtomicBatch.Put([]byte("GT"), gtBytes)

				if err := globals.BLOCKS.Write(blockDbAtomicBatch, nil); err != nil {
					panic("Can't store GT and block candidate")
				}

			}

		}

	}

}
