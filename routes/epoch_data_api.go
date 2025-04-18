package routes

import (
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/valyala/fasthttp"
)

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

func GetQuorumUrlsAndPubkeys(ctx *fasthttp.RequestCtx) {

}

func EpochProposition(ctx *fasthttp.RequestCtx) {

	// Handler for POST request

}
