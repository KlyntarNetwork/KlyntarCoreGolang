package life

import (
	"encoding/json"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
)

func timeIsOutForCurrentLeader(approvementThread *structures.ApprovementThread) bool {

	// Function to check if time frame for current leader is done and we have to move to next pool in sequence

	epochHandler := approvementThread.EpochHandler

	leaderShipTimeframe := approvementThread.NetworkParameters.LeadershipTimeframe

	currentIndex := int64(epochHandler.CurrentLeaderIndex)

	return utils.GetUTCTimestampInMilliSeconds() >= int64(epochHandler.StartTimestamp)+(currentIndex+1)*leaderShipTimeframe

}

func LeaderRotationThread() {

	globals.APPROVEMENT_THREAD_HANDLER.RWMutex.RLock()

	approvementThread := globals.APPROVEMENT_THREAD_HANDLER.Thread

	epochHandler := approvementThread.EpochHandler

	haveNextCandidate := epochHandler.CurrentLeaderIndex+1 < len(epochHandler.LeadersSequence)

	storedEpochIndex := epochHandler.Id

	if haveNextCandidate && timeIsOutForCurrentLeader(&approvementThread) {

		globals.APPROVEMENT_THREAD_HANDLER.RWMutex.RUnlock()

		globals.APPROVEMENT_THREAD_HANDLER.RWMutex.Lock()

		approvementThread = globals.APPROVEMENT_THREAD_HANDLER.Thread

		epochHandler = approvementThread.EpochHandler

		if storedEpochIndex == epochHandler.Id {

			approvementThread.EpochHandler.CurrentLeaderIndex++

			// Store the updated AT

			jsonedAT, _ := json.Marshal(approvementThread)

			globals.APPROVEMENT_THREAD_METADATA.Put([]byte("AT"), jsonedAT, nil)

		}

		globals.APPROVEMENT_THREAD_HANDLER.RWMutex.Unlock()

	} else {
		globals.APPROVEMENT_THREAD_HANDLER.RWMutex.RUnlock()
	}

	// The workflow of this function is infinite

	time.AfterFunc(200*time.Millisecond, func() {
		LeaderRotationThread()
	})

}
