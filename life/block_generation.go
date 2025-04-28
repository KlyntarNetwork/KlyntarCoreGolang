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
)

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

func generateBlocksPortion() {

}
