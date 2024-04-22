package tachyon

import (
	"fmt"

	klyGlobals "github.com/KLYN74R/KlyntarCoreGolang/KLY_Globals"
	tachyonLife "github.com/KLYN74R/KlyntarCoreGolang/KLY_Workflows/dev_tachyon/tachyon_life"
	"github.com/valyala/fasthttp"
)

func RunBlockchain() {

	PrepareBlockchain()

	//_________________________ RUN SEVERAL ASYNC THREADS _________________________

	//✅0.Start verification process - process blocks and find new epoch step-by-step
	go tachyonLife.StartVerificationThread()

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
	fasthttp.ListenAndServe(":8081", fastHTTPHandler)

}

// request handler in fasthttp style, i.e. just plain function.
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, klyGlobals.CONFIGS_PATH+"   => Hi there! RequestURI is %q", ctx.RequestURI())
}

func PrepareBlockchain() {

}

func CallBootStrapNodes() {

}
