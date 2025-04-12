package tachyon

import (
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/threads"
)

func getUtcTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

func IsMyCoreVersionOld(thread threads.ApprovementThread) bool {

	return thread.CoreMajorVersion > CORE_MAJOR_VERSION

}

func EpochStillFresh(thread threads.ApprovementThread) bool {

	return (thread.Epoch.StartTimestamp + uint64(thread.NetworkParameters.EpochTime)) > uint64(getUtcTimestamp())

}

type CurrentLeaderData struct {
	isMeLeader bool
	url        string
}

func GetCurrentLeader() CurrentLeaderData {

	return CurrentLeaderData{isMeLeader: false, url: ""}
}
