package routes

import (
	"fmt"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/valyala/fasthttp"
)

func GetFirstBlockAssumption(ctx *fasthttp.RequestCtx) {

	epochIndexVal := ctx.UserValue("epoch_index")
	epochIndex, ok := epochIndexVal.(string)

	if !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		ctx.Write([]byte(`{"err": "Invalid epoch index"}`))
		return
	}

	key := fmt.Sprintf("FIRST_BLOCK_ASSUMPTION:%s", epochIndex)
	value, err := globals.EPOCH_DATA.Get([]byte(key), nil)

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
