package life

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
	"github.com/gorilla/websocket"
)

type ProofsGrabber struct {
	EpochId        int
	AcceptedIndex  int
	AcceptedHash   string
	AfpForPrevious structures.AggregatedFinalizationProof
}

var WEBSOCKET_CONNECTIONS map[string]*websocket.Conn

var FINALIZATION_PROOFS_CACHE map[string]string

var RESPONSES chan Agreement

var PROOFS_GRABBER = ProofsGrabber{
	EpochId: -1,
}

func processIncomingFinalizationProof(msg []byte) {}

func runFinalizationProofsGrabbing() {

	// Call SendAndWait here
	// Once received 2/3 votes for block - continue

}

func BlocksSharingAndProofsGrabingThread() {

	globals.APPROVEMENT_THREAD.RWMutex.RLock()

	defer globals.APPROVEMENT_THREAD.RWMutex.RUnlock()

	epochHandler := globals.APPROVEMENT_THREAD.Thread.EpochHandler

	currentLeaderPubKey := epochHandler.LeadersSequence[epochHandler.CurrentLeaderIndex]

	if currentLeaderPubKey != globals.CONFIGURATION.PublicKey {

		time.AfterFunc(2*time.Second, func() {
			BlocksSharingAndProofsGrabingThread()
		})

		return

	}

	if PROOFS_GRABBER.EpochId != epochHandler.Id {

		// Try to get stored proofs grabber from db

		dbKey := []byte(strconv.Itoa(epochHandler.Id) + ":PROOFS_GRABBER")

		if rawGrabber, err := globals.FINALIZATION_VOTING_STATS.Get(dbKey, nil); err != nil {

			json.Unmarshal(rawGrabber, &PROOFS_GRABBER)

		} else {

			// Assign initial value of proofs grabber for each new epoch

			PROOFS_GRABBER = ProofsGrabber{

				EpochId: epochHandler.Id,

				AcceptedIndex: -1,

				AcceptedHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			}

		}

		// And store new descriptor

		if serialized, err := json.Marshal(PROOFS_GRABBER); err == nil {

			globals.FINALIZATION_VOTING_STATS.Put(dbKey, serialized, nil)

		}

		// Also, open connections with quorum here. Create QuorumWaiter etc.

		utils.OpenWebsocketConnectionsWithQuorum(epochHandler.Quorum, WEBSOCKET_CONNECTIONS)

	}

	// Continue here

	runFinalizationProofsGrabbing()

	go BlocksSharingAndProofsGrabingThread()

}
