package tachyon

import (
	"fmt"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/life"
	"github.com/valyala/fasthttp"
)

func RunBlockchain() {

	PrepareBlockchain()

	//_________________________ RUN SEVERAL THREADS _________________________

	//✅1.Thread to find AEFPs and change the epoch for AT
	go life.EpochRotationThread()

	//✅2.Share our blocks within quorum members and get the finalization proofs
	go life.BlocksSharingAndProofsGrabingThread()

	//✅3.Thread to propose AEFPs to move to next epoch
	go life.NewEpochProposerThread()

	//✅4.Start to generate blocks
	go life.BlocksGenerationThread()

	//✅5.Start a separate thread to work with voting for blocks in a sync way (for security)
	go life.VotingThread()

	// pass plain function to fasthttp

	// port := klyGlobals.CONFIGS["PORT"].(string)

	fasthttp.ListenAndServe(":8080", fastHTTPHandler)
}

// request handler in fasthttp style, i.e. just plain function.
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, CONFIGS_PATH+"   => Hi there! RequestURI is %q", ctx.RequestURI())
}

func PrepareBlockchain() {

}

func CallBootStrapNodes() {

}
