package life

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/system_contracts"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
)

type FirstBlockDataWithAefp struct {
	FirstBlockCreator, FirstBlockHash string

	Aefp *structures.AggregatedEpochFinalizationProof
}

var AEFP_AND_FIRST_BLOCK_DATA FirstBlockDataWithAefp

func ExecuteDelayedTransaction(originalTx []byte, parsedTx system_contracts.DelayedTransaction) {

	if funcHandler, ok := system_contracts.DELAYED_TRANSACTIONS_MAP[parsedTx.Type]; ok {

		funcHandler(originalTx)

	}

}

func fetchAefp(ctx context.Context, url string, quorum []string, majority int, epochFullID string, resultCh chan<- *structures.AggregatedEpochFinalizationProof) {

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var aefp *structures.AggregatedEpochFinalizationProof

	err = json.Unmarshal(body, aefp)

	if err == nil {

		if common_functions.VerifyAggregatedEpochFinalizationProof(aefp, quorum, majority, epochFullID) {

			select {

			case resultCh <- aefp:
			case <-ctx.Done():

			}

		}

	}

}

func EpochRotationThread() {

	if utils.EpochStillFresh(&globals.APPROVEMENT_THREAD) {

		epochHandler := globals.APPROVEMENT_THREAD.Epoch

		epochFullID := epochHandler.Hash + "#" + strconv.Itoa(epochHandler.Id)

		keyValue := []byte("EPOCH_FINISH_RESPONSE:" + strconv.Itoa(epochHandler.Id))

		readyToChangeEpochRaw, err := globals.FINALIZATION_VOTING_STATS.Get(keyValue, nil)

		if err == nil && string(readyToChangeEpochRaw) == "TRUE" {

			majority := common_functions.GetQuorumMajority(&epochHandler)

			quorumMembers := common_functions.GetQuorumUrlsAndPubkeys(&epochHandler)

			haveEverything := AEFP_AND_FIRST_BLOCK_DATA.Aefp != nil && AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash != ""

			if !haveEverything {

				// 1. Find AEFPs

				if AEFP_AND_FIRST_BLOCK_DATA.Aefp == nil {

					// Try to find locally first

					keyValue := []byte("AEFP:" + strconv.Itoa(epochHandler.Id))

					var aefp *structures.AggregatedEpochFinalizationProof

					aefpRaw, err := globals.EPOCH_DATA.Get(keyValue, nil)

					errParse := json.Unmarshal(aefpRaw, aefp)

					if err == nil && errParse == nil {

						AEFP_AND_FIRST_BLOCK_DATA.Aefp = aefp

					} else {

						// Ask quorum for AEFP

						resultCh := make(chan *structures.AggregatedEpochFinalizationProof, 1)

						ctx, cancel := context.WithCancel(context.Background())

						defer cancel() // safety net

						for _, quorumMember := range quorumMembers {

							go fetchAefp(ctx, quorumMember.Url, epochHandler.Quorum, majority, epochFullID, resultCh)

						}

						select {

						case value := <-resultCh:

							AEFP_AND_FIRST_BLOCK_DATA.Aefp = value
							cancel()

						case <-time.After(10 * time.Second):

							cancel()

						}

					}

				}

				// 2. Find first block in epoch

				if AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash == "" {

					firstBlockData := common_functions.GetFirstBlockInEpoch(&epochHandler)

					if firstBlockData != nil {

						AEFP_AND_FIRST_BLOCK_DATA.FirstBlockCreator = firstBlockData.FirstBlockCreator

						AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash = firstBlockData.FirstBlockHash

					}

				}

			}

			if AEFP_AND_FIRST_BLOCK_DATA.Aefp != nil && AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash != "" {

				firstBlock := common_functions.GetBlock(epochHandler.Id, AEFP_AND_FIRST_BLOCK_DATA.FirstBlockCreator, 0, &epochHandler)

				if firstBlock != nil && firstBlock.GetHash() == AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash {

					delayedTransactionsToExecute := [][]byte{}

					latestBatchIndexRaw, err := globals.APPROVEMENT_THREAD_METADATA.Get([]byte("LATEST_BATCH_INDEX"), nil)

					if delayedTransactionsBatchRaw, ok := firstBlock.ExtraData["delayedTxsBatch"]; ok {

					}

				}

			}

		}

		time.AfterFunc(0*time.Second, func() {
			EpochRotationThread()
		})

	} else {

		time.AfterFunc(3*time.Second, func() {
			EpochRotationThread()
		})

	}

}
