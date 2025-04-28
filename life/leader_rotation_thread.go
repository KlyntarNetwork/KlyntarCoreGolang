package life

import (
	"encoding/json"

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

	// TODO: Set .RLock() for RWMutex here

	approvementThread := globals.APPROVEMENT_THREAD

	epochHandler := approvementThread.EpochHandler

	haveNextCandidate := epochHandler.CurrentLeaderIndex+1 < len(epochHandler.LeadersSequence)

	if haveNextCandidate && timeIsOutForCurrentLeader(&approvementThread) {

		// Now, update the leader on approvement thread
		// TODO: Set .RUnlock() for RWMutex here
		// TODO: Set .Lock() for RWMutex here

		approvementThread.EpochHandler.CurrentLeaderIndex++

		// Store the updated AT

		jsonedAT, _ := json.Marshal(approvementThread)

		globals.APPROVEMENT_THREAD_METADATA.Put([]byte("AT"), jsonedAT, nil)

		// TODO: Release .Unlock() for RWMutex here

	}

}
