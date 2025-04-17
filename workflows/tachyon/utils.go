package tachyon

import (
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/structures"
)

func getUtcTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

func IsMyCoreVersionOld(thread *structures.ApprovementThread) bool {

	return thread.CoreMajorVersion > CORE_MAJOR_VERSION

}

func EpochStillFresh(thread *structures.ApprovementThread) bool {

	return (thread.Epoch.StartTimestamp + uint64(thread.NetworkParameters.EpochTime)) > uint64(getUtcTimestamp())

}

type CurrentLeaderData struct {
	isMeLeader bool
	url        string
}

func GetCurrentLeader() CurrentLeaderData {

	currentLeaderPubKey := APPROVEMENT_THREAD.Epoch.LeaderSequence[APPROVEMENT_THREAD.Epoch.CurrentLeaderIndex]

	if currentLeaderPubKey == CONFIGURATION.PublicKey {

		return CurrentLeaderData{isMeLeader: true, url: ""}

	}

	return CurrentLeaderData{isMeLeader: false, url: ""}
}
