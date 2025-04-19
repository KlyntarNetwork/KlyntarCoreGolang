package main

import (
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
)

type CurrentLeaderData struct {
	isMeLeader bool
	url        string
}

func getUtcTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

func IsMyCoreVersionOld(thread *structures.ApprovementThread) bool {

	return thread.CoreMajorVersion > globals.CORE_MAJOR_VERSION

}

func EpochStillFresh(thread *structures.ApprovementThread) bool {

	return (thread.Epoch.StartTimestamp + uint64(thread.NetworkParameters.EpochTime)) > uint64(getUtcTimestamp())

}

func GetCurrentLeader() CurrentLeaderData {

	currentLeaderPubKey := globals.APPROVEMENT_THREAD.Epoch.LeadersSequence[globals.APPROVEMENT_THREAD.Epoch.CurrentLeaderIndex]

	if currentLeaderPubKey == globals.CONFIGURATION.PublicKey {

		return CurrentLeaderData{isMeLeader: true, url: ""}

	}

	return CurrentLeaderData{isMeLeader: false, url: ""}
}
