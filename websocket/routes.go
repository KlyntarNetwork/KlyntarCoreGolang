package websocket

import (
	"encoding/json"
	"strconv"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/lxzan/gws"
)

func GetFinalizationProof(data any, connection *gws.Conn) {

	if parsedVal, ok := data.(WsFinalizationProofRequest); ok {

		epochHandler := globals.APPROVEMENT_THREAD.Epoch

		epochIndex := epochHandler.Id

		epochFullID := epochHandler.Hash + "#" + strconv.Itoa(epochIndex)

		currentLeaderIndex := epochHandler.CurrentLeaderIndex

		// conn.WriteMessage(gws.OpcodeText, []byte(`{"type":"pong"}`))

	}

}

func GetLeaderRotationProof(data any, connection *gws.Conn) {

	if parsedVal, ok := data.(WsLeaderRotationProofRequest); ok {

		epochHandler := globals.APPROVEMENT_THREAD.Epoch

		epochIndex := epochHandler.Id

		epochFullID := epochHandler.Hash + "#" + strconv.Itoa(epochIndex)

		currentLeaderIndex := epochHandler.CurrentLeaderIndex

		if epochHandler.CurrentLeaderIndex > parsedVal.HisIndexInLeadersSequence {

			localVotingData := structures.PoolVotingStat{
				Index: -1,
				Hash:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				Afp:   structures.AggregatedFinalizationProof{},
			}

			localVotingDataRaw, err := globals.FINALIZATION_VOTING_STATS.Get([]byte(strconv.Itoa(epochIndex) + ":" + parsedVal.PoolPubkey))

			if err == nil {

				json.Unmarshal(localVotingDataRaw, &localVotingData)

			}

			propSkipData := parsedVal.SkipData

			if localVotingData.Index > propSkipData.Index {

				// Try to return with AFP for the first block

				firstBlockID := ""

			}

		}

		// conn.WriteMessage(gws.OpcodeText, []byte(`{"type":"pong"}`))

	}

	// response, _ := json.Marshal(map[string]any{
	// 	"type": "echo",
	// 	"data": incoming.Data,
	// })
	// conn.WriteMessage(gws.OpcodeText, response)

}
