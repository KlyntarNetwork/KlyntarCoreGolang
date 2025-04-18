package main

import (
	"github.com/KlyntarNetwork/KlyntarCoreGolang/routes"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func NewRouter() fasthttp.RequestHandler {

	r := router.New()

	r.GET("/block/{id}", routes.GetBlockById)

	r.GET("/aggregated_finalization_proof/{blockId}", routes.GetAggregatedFinalizationProof)
	r.GET("/aggregated_epoch_finalization_proof/{epochIndex}", routes.GetAggregatedEpochFinalizationProof)

	r.GET("/first_block_assumption/{epochIndex}", routes.GetFirstBlockAssumption)

	// r.POST("/transaction")
	// r.POST("/epoch_proposition")

	return r.Handler
}
