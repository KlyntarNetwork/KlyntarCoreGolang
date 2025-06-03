package life

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
)

func timeIsOutForCurrentLeader(approvementThread *structures.ApprovementThreadMetadataHandler) bool {

	// Function to check if time frame for current leader is done and we have to move to next pool in sequence

	leaderShipTimeframe := approvementThread.NetworkParameters.LeadershipTimeframe

	currentIndex := int64(approvementThread.EpochHandler.CurrentLeaderIndex)

	return utils.GetUTCTimestampInMilliSeconds() >= int64(approvementThread.EpochHandler.StartTimestamp)+(currentIndex+1)*leaderShipTimeframe

}

func LeaderRotationThread() {

	for {

		globals.APPROVEMENT_THREAD_METADATA_HANDLER.RWMutex.RLock()

		epochHandler := &globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.EpochHandler

		haveNextCandidate := epochHandler.CurrentLeaderIndex+1 < len(epochHandler.LeadersSequence)

		if haveNextCandidate && timeIsOutForCurrentLeader(&globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler) {

			storedEpochIndex := epochHandler.Id

			globals.APPROVEMENT_THREAD_METADATA_HANDLER.RWMutex.RUnlock()

			globals.APPROVEMENT_THREAD_METADATA_HANDLER.RWMutex.Lock()

			threadHandler := &globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler

			if storedEpochIndex == threadHandler.EpochHandler.Id {

				threadHandler.EpochHandler.CurrentLeaderIndex++

				// Store the updated AT

				jsonedHandler, errMarshal := json.Marshal(threadHandler)

				if errMarshal != nil {

					fmt.Printf("Failed to marshal AT state: %v", errMarshal)

					panic("Impossible to marshal approvement thread state")

				}

				if err := globals.APPROVEMENT_THREAD_METADATA.Put([]byte("AT"), jsonedHandler, nil); err != nil {

					fmt.Printf("Failed to store AT state: %v", err)

					panic("Impossible to store the approvement thread state")

				}

			}

			globals.APPROVEMENT_THREAD_METADATA_HANDLER.RWMutex.Unlock()

		} else {
			globals.APPROVEMENT_THREAD_METADATA_HANDLER.RWMutex.RUnlock()
		}

		time.Sleep(200 * time.Millisecond)
	}

}
