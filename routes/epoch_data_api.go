package routes

import (
	"encoding/json"
	"fmt"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
	"github.com/valyala/fasthttp"
)

func sendJSON(ctx *fasthttp.RequestCtx, payload any) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	jsonBytes, _ := json.Marshal(payload)
	ctx.SetBody(jsonBytes)
}

func GetFirstBlockAssumption(ctx *fasthttp.RequestCtx) {

	epochIndexVal := ctx.UserValue("epochIndex")
	epochIndex, ok := epochIndexVal.(string)

	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		ctx.Write([]byte(`{"err": "Invalid epoch index"}`))
		return
	}

	value, err := globals.EPOCH_DATA.Get([]byte("FIRST_BLOCK_ASSUMPTION:"+epochIndex), nil)

	if err == nil && value != nil {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.Write(value)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetContentType("application/json")
	ctx.Write([]byte(`{"err": "No assumptions found"}`))
}

func GetAggregatedEpochFinalizationProof(ctx *fasthttp.RequestCtx) {

	epochIndexVal := ctx.UserValue("epochIndex")
	epochIndex, ok := epochIndexVal.(string)

	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		ctx.Write([]byte(`{"err": "Invalid epoch index"}`))
		return
	}

	value, err := globals.EPOCH_DATA.Get([]byte("AEFP:"+epochIndex), nil)

	if err == nil && value != nil {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.Write(value)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetContentType("application/json")
	ctx.Write([]byte(`{"err": "No assumptions found"}`))
}

func EpochProposition(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) != fasthttp.MethodPost {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		return
	}

	var proposition structures.EpochFinishProposition

	if err := json.Unmarshal(ctx.PostBody(), &proposition); err != nil {
		sendJSON(ctx, map[string]any{"err": "Wrong format"})
		return
	}

	// Todo: add mutex/atomic here for epochHandler (globals.APPROVEMENT_THREAD)

	epochHandler := globals.APPROVEMENT_THREAD.Epoch

	epochIndex := epochHandler.Id

	epochFullID := epochHandler.Hash + "#" + string(epochHandler.Id)

	localIndexOfLeader := epochHandler.CurrentLeaderIndex

	pubKeyOfCurrentLeader := epochHandler.LeadersSequence[localIndexOfLeader]

	signalRaw, err := globals.FINALIZATION_VOTING_STATS.Get([]byte("EPOCH_FINISH_RESPONSE:"+string(epochIndex)), nil)

	if err != nil || signalRaw == nil {
		sendJSON(ctx, map[string]any{"err": "Too early"})
		return
	}

	votingMetadataForPool := string(epochIndex) + ":" + pubKeyOfCurrentLeader

	votingRaw, err := globals.FINALIZATION_VOTING_STATS.Get([]byte(votingMetadataForPool), nil)

	var votingData structures.PoolVotingStat

	if err != nil || votingRaw == nil {

		votingData = structures.PoolVotingStat{
			Index: -1,
			Hash:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			Afp:   structures.AggregatedFinalizationProof{},
		}

	} else {
		_ = json.Unmarshal(votingRaw, &votingData)
	}

	blockID := string(epochHandler.Id) + ":" + pubKeyOfCurrentLeader + ":0"

	var hashOfFirstBlock string

	if proposition.AfpForFirstBlock.BlockID == blockID && proposition.LastBlockProposition.Index >= 0 {

		if common_functions.VerifyAggregatedFinalizationProof(&proposition.AfpForFirstBlock, &epochHandler) {

			hashOfFirstBlock = proposition.AfpForFirstBlock.BlockHash

		}

	}

	if hashOfFirstBlock == "" {

		sendJSON(ctx, map[string]any{"err": "Can't verify hash"})

		return

	}

	response := map[string]any{}

	if proposition.CurrentLeader == localIndexOfLeader {

		if votingData.Index == proposition.LastBlockProposition.Index && votingData.Hash == proposition.LastBlockProposition.Hash {

			dataToSign := fmt.Sprintf("EPOCH_DONE:%d:%d:%s:%s:%s",
				proposition.CurrentLeader,
				proposition.LastBlockProposition.Index,
				proposition.LastBlockProposition.Hash,
				hashOfFirstBlock,
				epochFullID,
			)

			sig := ed25519.GenerateSignature(globals.CONFIGURATION.PrivateKey, dataToSign)

			response = map[string]any{
				"status": "OK",
				"sig":    sig,
			}

		} else if votingData.Index > proposition.LastBlockProposition.Index {

			response = map[string]any{
				"status":               "UPGRADE",
				"currentLeader":        localIndexOfLeader,
				"lastBlockProposition": votingData,
			}

		}

	} else if proposition.CurrentLeader < localIndexOfLeader {

		response = map[string]any{
			"status":               "UPGRADE",
			"currentLeader":        localIndexOfLeader,
			"lastBlockProposition": votingData,
		}

	}

	sendJSON(ctx, response)
}
