package life

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
)

var CURRENT_LEADER_INDEX = 0 // we need this value "globally" as a cache

var QUORUM_AGREEMENTS map[string]string

type ResponseStatus struct {
	Status string
}

func NewEpochProposerThread() {

	globals.APPROVEMENT_THREAD.RWMutex.RLock()

	if !utils.EpochStillFresh(&globals.APPROVEMENT_THREAD.Thread) {

		globals.APPROVEMENT_THREAD.RWMutex.RUnlock()

		globals.APPROVEMENT_THREAD.RWMutex.Lock()

		defer globals.APPROVEMENT_THREAD.RWMutex.Unlock()

		atEpochHandler := globals.APPROVEMENT_THREAD.Thread.EpochHandler

		epochIndex := atEpochHandler.Id

		epochFullID := atEpochHandler.Hash + "#" + strconv.Itoa(atEpochHandler.Id)

		leadersSequence := atEpochHandler.LeadersSequence

		pubKeyOfLeader := leadersSequence[CURRENT_LEADER_INDEX]

		iAmInTheQuorum := slices.Contains(atEpochHandler.Quorum, globals.CONFIGURATION.PublicKey)

		if iAmInTheQuorum {

			majority := common_functions.GetQuorumMajority(&atEpochHandler)

			var localVotingData structures.PoolVotingStat

			localVotingDataRaw, err := globals.FINALIZATION_VOTING_STATS.Get([]byte(strconv.Itoa(epochIndex)+":"+pubKeyOfLeader), nil)

			if err != nil {

				localVotingData = structures.PoolVotingStat{
					Index: -1,
					Hash:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
					Afp:   structures.AggregatedFinalizationProof{},
				}

			} else {

				json.Unmarshal(localVotingDataRaw, &localVotingData)

			}

			if localVotingData.Index == -1 {

				for position := CURRENT_LEADER_INDEX - 1; position >= 0; position-- {

					prevLeader := atEpochHandler.LeadersSequence[position]

					prevVotingDataRaw, err := globals.FINALIZATION_VOTING_STATS.Get([]byte(strconv.Itoa(epochIndex)+":"+prevLeader), nil)

					if err == nil {

						var prevVotingData structures.PoolVotingStat

						json.Unmarshal(prevVotingDataRaw, &prevVotingData)

						if prevVotingData.Index > -1 {

							pubKeyOfLeader = prevLeader

							CURRENT_LEADER_INDEX = position

							localVotingData = prevVotingData

							break

						}

					}

				}

			}

			var epochFinishProposition structures.EpochFinishRequest

			if _, err := globals.EPOCH_DATA.Get([]byte("AEFP:"+strconv.Itoa(epochIndex)), nil); err != nil {

				firstBlockID := strconv.Itoa(epochIndex) + ":" + pubKeyOfLeader + ":0"

				afpForFirstBlockRaw, _ := globals.EPOCH_DATA.Get([]byte("AFP:"+firstBlockID), nil)

				var afpForFirstBlock structures.AggregatedFinalizationProof

				json.Unmarshal(afpForFirstBlockRaw, &afpForFirstBlock)

				epochFinishProposition = structures.EpochFinishRequest{
					CurrentLeader:        CURRENT_LEADER_INDEX,
					LastBlockProposition: localVotingData,
					AfpForFirstBlock:     afpForFirstBlock,
				}
			}

			quorumMembers := common_functions.GetQuorumUrlsAndPubkeys(&atEpochHandler)

			resultsCh := make(chan Agreement, len(quorumMembers))
			upgradeCh := make(chan structures.EpochFinishResponseUpgrade, len(quorumMembers))

			var wg sync.WaitGroup

			for _, descriptor := range quorumMembers {

				if _, ok := QUORUM_AGREEMENTS[descriptor.PubKey]; ok {
					continue
				}

				wg.Add(1)

				go func(desc common_functions.QuorumMemberData) {

					defer wg.Done()

					body, _ := json.Marshal(epochFinishProposition)

					ctxReq, cancel := context.WithTimeout(context.Background(), 3*time.Second)

					defer cancel()

					req, _ := http.NewRequestWithContext(ctxReq, "POST", desc.Url+"/epoch_proposition", bytes.NewReader(body))

					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}

					resp, err := client.Do(req)

					if err != nil {
						return
					}

					defer resp.Body.Close()

					responseBytes, err := io.ReadAll(resp.Body)

					if err != nil {
						return
					}

					var respStatus ResponseStatus

					if err := json.Unmarshal(responseBytes, &respStatus); err != nil {
						return
					}

					switch respStatus.Status {

					case "OK":

						dataToSign := strconv.Itoa(epochFinishProposition.CurrentLeader) + ":" +
							strconv.Itoa(epochFinishProposition.LastBlockProposition.Index) + ":" +
							epochFinishProposition.LastBlockProposition.Hash + ":" +
							epochFinishProposition.AfpForFirstBlock.BlockHash + ":" +
							epochFullID

						var resultAsStruct structures.EpochFinishResponseOk

						json.Unmarshal(responseBytes, &resultAsStruct)

						if ed25519.VerifySignature(dataToSign, desc.PubKey, resultAsStruct.Sig) {

							resultsCh <- Agreement{
								PubKey: desc.PubKey,
								Sig:    resultAsStruct.Sig,
							}

						}

					case "UPGRADE":

						var resultAsStruct structures.EpochFinishResponseUpgrade

						json.Unmarshal(responseBytes, &resultAsStruct)

						if common_functions.VerifyAggregatedFinalizationProof(&resultAsStruct.LastBlockProposition.Afp, &atEpochHandler) {

							blockID := strconv.Itoa(epochIndex) + ":" +
								leadersSequence[resultAsStruct.CurrentLeader] + ":" +
								strconv.Itoa(resultAsStruct.LastBlockProposition.Index)

							sameBlockID := blockID == resultAsStruct.LastBlockProposition.Afp.BlockID

							sameHash := resultAsStruct.LastBlockProposition.Hash == resultAsStruct.LastBlockProposition.Afp.BlockHash

							proposedLeaderHasBiggerIndex := resultAsStruct.CurrentLeader > CURRENT_LEADER_INDEX

							if sameBlockID && sameHash && proposedLeaderHasBiggerIndex {

								upgradeCh <- resultAsStruct

							}

						}

					}

				}(descriptor)

			}

			go func() {
				wg.Wait()
				close(resultsCh)
				close(upgradeCh)
			}()

			for result := range resultsCh {

				QUORUM_AGREEMENTS[result.PubKey] = result.Sig

			}

			for upgradeProposition := range upgradeCh {

				if upgradeProposition.CurrentLeader > CURRENT_LEADER_INDEX {

					CURRENT_LEADER_INDEX = upgradeProposition.CurrentLeader

					keyAsBytes := []byte(strconv.Itoa(epochIndex) + ":" + leadersSequence[CURRENT_LEADER_INDEX])

					valueAsBytes, _ := json.Marshal(upgradeProposition.LastBlockProposition)

					globals.FINALIZATION_VOTING_STATS.Put(keyAsBytes, valueAsBytes, nil)

					// In this case - clear the quorum agreements to try grab proofs for leader with bigger index

					clear(QUORUM_AGREEMENTS)

				}

			}

			if len(QUORUM_AGREEMENTS) >= majority {

				aggregatedEpochFinalizationProof := structures.AggregatedEpochFinalizationProof{
					LastLeader:                   uint(epochFinishProposition.CurrentLeader),
					LastIndex:                    uint(epochFinishProposition.LastBlockProposition.Index),
					LastHash:                     epochFinishProposition.LastBlockProposition.Hash,
					HashOfFirstBlockByLastLeader: epochFinishProposition.AfpForFirstBlock.BlockHash,
					Proofs:                       QUORUM_AGREEMENTS,
				}

				if common_functions.VerifyAggregatedEpochFinalizationProof(&aggregatedEpochFinalizationProof, atEpochHandler.Quorum, majority, epochFullID) {

					valueAsBytes, _ := json.Marshal(aggregatedEpochFinalizationProof)

					globals.EPOCH_DATA.Put([]byte("AEFP:"+strconv.Itoa(epochIndex)), valueAsBytes, nil)

				}

			}

		}

	} else {

		globals.APPROVEMENT_THREAD.RWMutex.RUnlock()

		time.AfterFunc(3*time.Second, func() {
			NewEpochProposerThread()
		})

	}

}
