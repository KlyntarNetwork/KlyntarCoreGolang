package main

import (
	"log"
	"strconv"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/life"
	"github.com/valyala/fasthttp"
)

func RunBlockchain() {

	prepareBlockchain()

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

	serverAddr := globals.CONFIGURATION.Interface + ":" + strconv.Itoa(globals.CONFIGURATION.Port)

	err := fasthttp.ListenAndServe(serverAddr, NewRouter())

	if err != nil {
		log.Fatalf("Error in server: %s", err)
	}

}

func prepareBlockchain() {

}
