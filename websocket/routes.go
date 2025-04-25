package websocket

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
	"github.com/lxzan/gws"
)

func IndexOf(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}
	return -1
}

func GetFinalizationProof(data any, connection *gws.Conn) {

	if parsedRequest, ok := data.(WsFinalizationProofRequest); ok {

		epochHandler := globals.APPROVEMENT_THREAD.Epoch

		epochIndex := epochHandler.Id

		epochFullID := epochHandler.Hash + "#" + strconv.Itoa(epochIndex)

		typeCheckIsOk := &parsedRequest.Block != nil && &parsedRequest.PreviousBlockAfp != nil

		itsLeader := epochHandler.LeadersSequence[epochHandler.CurrentLeaderIndex] == parsedRequest.Block.Creator

		if typeCheckIsOk && itsLeader {

			localVotingDataForPool := structures.PoolVotingStat{
				Index: -1,
				Hash:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Afp:   structures.AggregatedFinalizationProof{},
			}

			localVotingDataRaw, err := globals.FINALIZATION_VOTING_STATS.Get([]byte(strconv.Itoa(epochIndex)+":"+parsedRequest.Block.Creator), nil)

			if err == nil {

				json.Unmarshal(localVotingDataRaw, &localVotingDataForPool)

			}

			proposedBlockHash := parsedRequest.Block.GetHash()

			itsSameChainSegment := localVotingDataForPool.Index < int(parsedRequest.Block.Index) || localVotingDataForPool.Index == int(parsedRequest.Block.Index) && proposedBlockHash == localVotingDataForPool.Hash && parsedRequest.Block.Epoch == epochFullID

			if itsSameChainSegment {

				proposedBlockId := strconv.Itoa(epochIndex) + ":" + parsedRequest.Block.Creator + strconv.Itoa(int(parsedRequest.Block.Index))

				futureVotingDataToStore := structures.PoolVotingStat{}

				if parsedRequest.Block.VerifySignature() {

					if localVotingDataForPool.Index == int(parsedRequest.Block.Index) {

						futureVotingDataToStore = localVotingDataForPool

					} else {

						futureVotingDataToStore = structures.PoolVotingStat{

							Index: int(parsedRequest.Block.Index - 1),

							Hash: parsedRequest.PreviousBlockAfp.BlockHash,

							Afp: parsedRequest.PreviousBlockAfp,
						}

					}

					previousBlockID := ""

					if parsedRequest.Block.Index == 0 {

						aefpIsOk := false

						if epochIndex == 0 {

							aefpIsOk = true

						} else {

							var legacyEpochHandler *structures.EpochHandler

							prevEpochIndex := epochHandler.Id - 1

							legacyEpochHandlerRaw, err := globals.EPOCH_DATA.Get([]byte("EPOCH_HANDLER:"+strconv.Itoa(prevEpochIndex)), nil)

							if err == nil {

								errParse := json.Unmarshal(legacyEpochHandlerRaw, legacyEpochHandler)

								aefpFromBlock, typeAssertOk := parsedRequest.Block.ExtraData["aefpForPreviousEpoch"].(structures.AggregatedEpochFinalizationProof)

								if typeAssertOk && errParse == nil {

									legacyEpochFullID := legacyEpochHandler.Hash + "#" + strconv.Itoa(legacyEpochHandler.Id)

									legacyMajority := common_functions.GetQuorumMajority(legacyEpochHandler)

									aefpIsOk = epochHandler.Id == 0 || common_functions.VerifyAggregatedEpochFinalizationProof(

										&aefpFromBlock,

										legacyEpochHandler.Quorum,

										legacyMajority,

										legacyEpochFullID,
									)

								}

							}

						}

						//_________________________________________2_________________________________________

						// TODO: Verify the ALRP chain validity here

					} else {

					}

				}

			}

		}

		// conn.WriteMessage(gws.OpcodeText, []byte(`{"type":"pong"}`))

	}

}

func GetLeaderRotationProof(data any, connection *gws.Conn) {

	if parsedRequest, typeAssertOk := data.(WsLeaderRotationProofRequest); typeAssertOk {

		epochHandler := globals.APPROVEMENT_THREAD.Epoch

		epochIndex := epochHandler.Id

		epochFullID := epochHandler.Hash + "#" + strconv.Itoa(epochIndex)

		poolToRotate := epochHandler.LeadersSequence[parsedRequest.IndexOfPoolToRotate]

		if epochHandler.CurrentLeaderIndex > parsedRequest.IndexOfPoolToRotate {

			localVotingData := structures.PoolVotingStat{
				Index: -1,
				Hash:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Afp:   structures.AggregatedFinalizationProof{},
			}

			localVotingDataRaw, err := globals.FINALIZATION_VOTING_STATS.Get([]byte(strconv.Itoa(epochIndex)+":"+poolToRotate), nil)

			if err == nil {

				json.Unmarshal(localVotingDataRaw, &localVotingData)

			}

			propSkipData := parsedRequest.SkipData

			if localVotingData.Index > propSkipData.Index {

				// Try to return with AFP for the first block

				firstBlockID := fmt.Sprintf("%d:%s:0", epochHandler.Id, poolToRotate)

				afpForFirstBlockBytes, err := globals.EPOCH_DATA.Get([]byte("AFP:"+firstBlockID), nil)

				if err == nil {

					var afpForFirstBlock structures.AggregatedFinalizationProof

					err := json.Unmarshal(afpForFirstBlockBytes, &afpForFirstBlock)

					if err == nil {

						responseData := WsLeaderRotationProofResponseUpgrade{

							Route:            "get_leader_rotation_proof",
							Voter:            globals.CONFIGURATION.PublicKey,
							ForPoolPubkey:    poolToRotate,
							Type:             "UPDATE",
							AfpForFirstBlock: afpForFirstBlock,
							SkipData:         localVotingData,
						}

						jsonResponse, err := json.Marshal(responseData)

						if err == nil {

							connection.WriteMessage(gws.OpcodeText, jsonResponse)

						}

					}

				}

			} else {

				//________________________________________________ Verify the proposed AFP ________________________________________________

				afpIsOk := false

				parts := strings.Split(propSkipData.Afp.BlockID, ":")

				if len(parts) != 3 {
					return
				}

				indexOfBlockInAfp, err := strconv.Atoi(parts[2])

				if err != nil {
					return
				}

				if propSkipData.Index > -1 && propSkipData.Hash == propSkipData.Afp.BlockHash && propSkipData.Index == indexOfBlockInAfp {

					afpIsOk = common_functions.VerifyAggregatedFinalizationProof(&propSkipData.Afp, &epochHandler)

				} else {

					afpIsOk = true
				}

				if afpIsOk {

					dataToSignForLeaderRotation, firstBlockAfpIsOk := "", false

					if parsedRequest.SkipData.Index == -1 {

						dataToSignForLeaderRotation = fmt.Sprintf(
							"LEADER_ROTATION_PROOF:%s:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:-1:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:%s",
							poolToRotate,
							epochFullID,
						)

						firstBlockAfpIsOk = true

					} else if parsedRequest.SkipData.Index >= 0 && &parsedRequest.AfpForFirstBlock != nil {

						blockIdOfFirstBlock := strconv.Itoa(epochIndex) + ":" + poolToRotate + ":0"

						blockIdsTheSame := parsedRequest.AfpForFirstBlock.BlockID == blockIdOfFirstBlock

						if blockIdsTheSame && common_functions.VerifyAggregatedFinalizationProof(&parsedRequest.AfpForFirstBlock, &epochHandler) {

							firstBlockHash := parsedRequest.AfpForFirstBlock.BlockHash

							dataToSignForLeaderRotation = fmt.Sprintf(
								"LEADER_ROTATION_PROOF:%s:%s:%d:%s:%s",
								poolToRotate,
								firstBlockHash,
								propSkipData.Index,
								propSkipData.Hash,
								epochFullID,
							)

							firstBlockAfpIsOk = true

						}

					}

					// If proof is ok - generate LRP(leader rotation proof)

					if firstBlockAfpIsOk {

						leaderRotationProofMessage := WsLeaderRotationProofResponseOk{

							Route: "get_leader_rotation_proof",

							Voter: globals.CONFIGURATION.PublicKey,

							ForPoolPubkey: poolToRotate,

							Type: "OK",

							Sig: ed25519.GenerateSignature(globals.CONFIGURATION.PrivateKey, dataToSignForLeaderRotation),
						}

						jsonResponse, err := json.Marshal(leaderRotationProofMessage)

						if err == nil {

							connection.WriteMessage(gws.OpcodeText, jsonResponse)

						}

					}

				}

			}

		}

	}

}

func GetBlocksRange(data any, connection *gws.Conn) {

}
