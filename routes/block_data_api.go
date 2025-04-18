package routes

import (
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/valyala/fasthttp"
)

func GetBlockById(ctx *fasthttp.RequestCtx) {

	blockIdRaw := ctx.UserValue("id")
	blockId, ok := blockIdRaw.(string)

	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		ctx.Write([]byte(`{"err": "Invalid value"}`))
		return
	}

	block, err := globals.BLOCKS.Get([]byte(blockId), nil)

	if err == nil && block != nil {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.Write(block)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetContentType("application/json")
	ctx.Write([]byte(`{"err": "Not found"}`))
}

func GetAggregatedFinalizationProof(ctx *fasthttp.RequestCtx) {

	blockIdRaw := ctx.UserValue("blockId")
	blockId, ok := blockIdRaw.(string)

	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		ctx.Write([]byte(`{"err": "Invalid value"}`))
		return
	}

	afp, err := globals.EPOCH_DATA.Get([]byte("AFP:"+blockId), nil)

	if err == nil && afp != nil {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.Write(afp)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetContentType("application/json")
	ctx.Write([]byte(`{"err": "Not found"}`))
}

func AcceptTransaction(ctx *fasthttp.RequestCtx) {

	// Handler for POST request

}
