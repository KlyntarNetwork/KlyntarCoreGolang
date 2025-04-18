package tachyon

import (
	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/routes"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func NewRouter() fasthttp.RequestHandler {

	r := router.New()

	r.GET("/first_block_assumption/{epoch_index}", routes.GetFirstBlockAssumption)

	return r.Handler
}
