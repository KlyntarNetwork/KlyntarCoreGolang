package tachyon

import (
	"fmt"

	klyGlobals "github.com/KlyntarNetwork/KlyntarCoreGolang/KLY_Globals"
	tachyonLife "github.com/KlyntarNetwork/KlyntarCoreGolang/KLY_Workflows/dev_tachyon/tachyon_life"
	"github.com/valyala/fasthttp"
)

func RunBlockchain() {

	PrepareBlockchain()

	//_________________________ RUN SEVERAL THREADS _________________________

	//✅1.Thread to find AEFPs and change the epoch for QT
	go tachyonLife.FindAggregatedEpochFinalizationProofs()

	//✅2.Share our blocks within quorum members and get the finalization proofs
	go tachyonLife.ShareBlocksAndGetFinalizationProofs()

	//✅3.Thread to propose AEFPs to move to next epoch
	go tachyonLife.CheckIfItsTimeToStartNewEpoch()

	//✅4.Thread to track changes of leaders on shards
	go tachyonLife.ShardsLeadersMonitoring()

	//✅5.Function to build the temporary sequence of blocks to verify them
	go tachyonLife.BuildTemporarySequenceForVerificationThread()

	//✅6.Start to generate blocks
	go tachyonLife.BlockGeneration()

	// pass plain function to fasthttp

	port := klyGlobals.CONFIGS["PORT"].(string)

	fasthttp.ListenAndServe(":"+port, fastHTTPHandler)
}

// request handler in fasthttp style, i.e. just plain function.
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, klyGlobals.CONFIGS_PATH+"   => Hi there! RequestURI is %q", ctx.RequestURI())
}

func PrepareBlockchain() {

}

func CallBootStrapNodes() {

}
