package life

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/system_contracts"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
	"github.com/syndtr/goleveldb/leveldb"
)

type FirstBlockDataWithAefp struct {
	FirstBlockCreator, FirstBlockHash string

	Aefp *structures.AggregatedEpochFinalizationProof
}

var AEFP_AND_FIRST_BLOCK_DATA FirstBlockDataWithAefp

func ExecuteDelayedTransaction(delayedTransaction map[string]string) {

	if delayedTxType, ok := delayedTransaction["type"]; ok {

		// Now find the handler

		if funcHandler, ok := system_contracts.DELAYED_TRANSACTIONS_MAP[delayedTxType]; ok {

			funcHandler(delayedTransaction)

		}

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

		epochHandler := globals.APPROVEMENT_THREAD.EpochHandler

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

				// 1. Fetch first block

				firstBlock := common_functions.GetBlock(epochHandler.Id, AEFP_AND_FIRST_BLOCK_DATA.FirstBlockCreator, 0, &epochHandler)

				// 2. Compare hashes

				if firstBlock != nil && firstBlock.GetHash() == AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash {

					// 3. Verify that quorum agreed batch of delayed transactions

					latestBatchIndex := int64(0)

					latestBatchIndexRaw, err := globals.APPROVEMENT_THREAD_METADATA.Get([]byte("LATEST_BATCH_INDEX"), nil)

					if err == nil {

						latestBatchIndex = utils.BytesToInt(latestBatchIndexRaw)

					}

					var delayedTransactionsToExecute []map[string]string

					if delayedTransactionsBatchRaw, ok := firstBlock.ExtraData["delayedTxsBatch"]; ok {

						// Convert to json first

						jsonData, err := json.Marshal(delayedTransactionsBatchRaw)

						if err == nil {

							var delayedTransactionsBatch system_contracts.DelayedTransactionsBatch

							if err := json.Unmarshal(jsonData, &delayedTransactionsBatch); err == nil {

								// 4. Verify signatures first

								jsonedDelayedTxs, _ := json.Marshal(delayedTransactionsBatch.DelayedTransactions)

								dataThatShouldBeSigned := "SIG_DELAYED_OPERATIONS:" +
									strconv.Itoa(epochHandler.Id) + ":" +
									string(jsonedDelayedTxs)

								okSignatures := 0
								unique := make(map[string]bool)
								quorumMap := make(map[string]bool)

								for _, pk := range epochHandler.Quorum {
									quorumMap[strings.ToLower(pk)] = true
								}

								for signerPubKey, signa := range delayedTransactionsBatch.Proofs {

									isOK := ed25519.VerifySignature(dataThatShouldBeSigned, signerPubKey, signa)

									loweredPubKey := strings.ToLower(signerPubKey)

									if isOK && quorumMap[signerPubKey] && !unique[loweredPubKey] {

										unique[loweredPubKey] = true
										okSignatures++

									}

								}

								if okSignatures >= majority {
									// 5. Finally - check if this batch has bigger index than already executed
									// 6. Only in case it's indeed new batch - execute it
									if int64(epochHandler.Id) > latestBatchIndex {

										latestBatchIndex = int64(epochHandler.Id)
										delayedTransactionsToExecute = delayedTransactionsBatch.DelayedTransactions

									}

								}

							}

						}

					}

					keyBytes := []byte("EPOCH_HANDLER:" + strconv.Itoa(epochHandler.Id))

					valBytes, _ := json.Marshal(epochHandler)

					globals.EPOCH_DATA.Put(keyBytes, valBytes, nil)

					var daoVotingContractCalls []map[string]string
					var allTheRestContractCalls []map[string]string

					atomicBatch := new(leveldb.Batch)

					for _, delayedTransaction := range delayedTransactionsToExecute {

						if delayedTxType, ok := delayedTransaction["type"]; ok {

							if delayedTxType == "votingAccept" {
								daoVotingContractCalls = append(daoVotingContractCalls, delayedTransaction)
							} else {
								allTheRestContractCalls = append(allTheRestContractCalls, delayedTransaction)
							}

						}

					}

					delayedTransactionsOrderByPriority := append(daoVotingContractCalls, allTheRestContractCalls...)

					// Execute delayed transactions
					for _, delayedTransaction := range delayedTransactionsOrderByPriority {

						ExecuteDelayedTransaction(delayedTransaction)

					}

					for key, value := range globals.APPROVEMENT_THREAD.Cache {

						valBytes, _ := json.Marshal(value)

						atomicBatch.Put([]byte(key), valBytes)

					}

					utils.LogWithTime("Dealyed txs were executed for epoch on AT: "+epochFullID, utils.GREEN_COLOR)

					//_______________________ Update the values for new epoch _______________________

					// Now, after the execution we can change the epoch id and get the new hash + prepare new temporary object

					nextEpochId := epochHandler.Id + 1

					nextEpochHash := utils.Blake3(AEFP_AND_FIRST_BLOCK_DATA.FirstBlockHash)

					nextEpochQuorumSize := globals.APPROVEMENT_THREAD.NetworkParameters.QuorumSize

					nextEpochHandler := structures.EpochHandler{
						Id:                 nextEpochId,
						Hash:               nextEpochHash,
						PoolsRegistry:      epochHandler.PoolsRegistry,
						ShardsRegistry:     epochHandler.ShardsRegistry,
						Quorum:             common_functions.GetCurrentEpochQuorum(&epochHandler, nextEpochQuorumSize, nextEpochHash),
						LeadersSequence:    []string{},
						StartTimestamp:     epochHandler.StartTimestamp + uint64(globals.APPROVEMENT_THREAD.NetworkParameters.EpochTime),
						CurrentLeaderIndex: 0,
					}

					common_functions.SetLeadersSequence(&nextEpochHandler, nextEpochHash)

					atomicBatch.Put([]byte("LATEST_BATCH_INDEX:"), []byte(strconv.Itoa(int(latestBatchIndex))))

					globals.APPROVEMENT_THREAD.EpochHandler = nextEpochHandler

					jsonedAT, _ := json.Marshal(globals.APPROVEMENT_THREAD)

					atomicBatch.Put([]byte("AT"), jsonedAT)

					// Clean cache

					clear(globals.APPROVEMENT_THREAD.Cache)

					globals.APPROVEMENT_THREAD_METADATA.Write(atomicBatch, nil)

					utils.LogWithTime("Epoch on approvement thread was updated => "+nextEpochHash+"#"+strconv.Itoa(nextEpochId), utils.GREEN_COLOR)

					//_______________________Check the version required for the next epoch________________________

					if utils.IsMyCoreVersionOld(&globals.APPROVEMENT_THREAD) {

						utils.LogWithTime("New version detected on APPROVEMENT_THREAD. Please, upgrade your node software", utils.YELLOW_COLOR)

						utils.GracefulShutdown()

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
