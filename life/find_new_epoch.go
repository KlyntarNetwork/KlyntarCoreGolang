package life

import (
	"encoding/json"
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
