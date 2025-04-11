package tachyon

import (
	"fmt"

	klyGlobals "github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	tachyonLife "github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/life"
	"github.com/valyala/fasthttp"
)

func RunBlockchain() {

	PrepareBlockchain()

	//_________________________ RUN SEVERAL THREADS _________________________

	//✅1.Thread to find AEFPs and change the epoch for AT
	go tachyonLife.EpochRotationThread()

	//✅2.Share our blocks within quorum members and get the finalization proofs
	go tachyonLife.BlocksSharingAndProofsGrabingThread()

	//✅3.Thread to propose AEFPs to move to next epoch
	go tachyonLife.NewEpochProposerThread()

	//✅4.Thread to track changes of leaders on shards
	go tachyonLife.LeadersSequenceeMonitoring()

	//✅5.Function to build the temporary sequence of blocks to verify them
	go tachyonLife.BlocksOrderingForExecutionThread()

	//✅6.Start to generate blocks
	go tachyonLife.BlocksGenerationThread()

	//✅7.Start a separate thread to work with voting for blocks in a sync way (for security)
	go tachyonLife.VotingThread()

	// pass plain function to fasthttp

	// port := klyGlobals.CONFIGS["PORT"].(string)

	fasthttp.ListenAndServe(":8080", fastHTTPHandler)
}

// request handler in fasthttp style, i.e. just plain function.
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, klyGlobals.CONFIGS_PATH+"   => Hi there! RequestURI is %q", ctx.RequestURI())
}

func PrepareBlockchain() {

}

func CallBootStrapNodes() {

}
