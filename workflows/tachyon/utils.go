package tachyon

import (
	"math/rand"

	"time"

	threads "github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/threads"
)

func getUtcTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

func GetRandomFromArray(arr []string) string {

	randomIndex := rand.Intn(len(arr))

	return arr[randomIndex]

}

func EpochStillFresh(thread threads.ApprovementThread) bool {

	return (thread.Epoch.StartTimestamp + uint64(thread.NetworkParameters.EpochTime)) > uint64(getUtcTimestamp())

}
