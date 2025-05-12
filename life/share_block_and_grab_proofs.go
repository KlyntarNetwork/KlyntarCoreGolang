package life

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/block"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
	"github.com/gorilla/websocket"
)

type ProofsGrabber struct {
	EpochId             int
	AcceptedIndex       int
	AcceptedHash        string
	AfpForPrevious      structures.AggregatedFinalizationProof
	HuntingForBlockId   string
	HuntingForBlockHash string
}

var WEBSOCKET_CONNECTIONS map[string]*websocket.Conn

var FINALIZATION_PROOFS_CACHE map[string]string

var PROOFS_GRABBER = ProofsGrabber{
	EpochId: -1,
}

var BLOCK_TO_SHARE *block.Block

func processIncomingFinalizationProof(msg []byte) {}

func runFinalizationProofsGrabbing() {

	// Call SendAndWait here
	// Once received 2/3 votes for block - continue

	epochHandler := globals.APPROVEMENT_THREAD.Thread.EpochHandler

	blockIndexToHunt := strconv.Itoa(PROOFS_GRABBER.AcceptedIndex + 1)

	blockIdForHunting := strconv.Itoa(epochHandler.Id) + ":" + globals.CONFIGURATION.PublicKey + ":" + blockIndexToHunt

	majority := common_functions.GetQuorumMajority(&epochHandler)

	if BLOCK_TO_SHARE == nil {

		// Get from db and assign. If no such block - return

	}

	blockHash := BLOCK_TO_SHARE.GetHash()

	PROOFS_GRABBER.HuntingForBlockId = blockIdForHunting

	PROOFS_GRABBER.HuntingForBlockHash = blockHash

	if len(FINALIZATION_PROOFS_CACHE) < majority {

		// Initiate request

	}

	if len(FINALIZATION_PROOFS_CACHE) >= majority {

		aggregatedFinalizationProof := structures.AggregatedFinalizationProof{

			PrevBlockHash: PROOFS_GRABBER.AcceptedHash,

			BlockID: blockIdForHunting,

			BlockHash: blockHash,

			Proofs: FINALIZATION_PROOFS_CACHE,
		}

		keyBytes := []byte("AFP:" + blockIdForHunting)

		valueBytes, _ := json.Marshal(aggregatedFinalizationProof)

		// Store locally
		globals.EPOCH_DATA.Put(keyBytes, valueBytes, nil)

		// Delete finalization proofs that we don't need more
		FINALIZATION_PROOFS_CACHE = map[string]string{}

		// Repeat procedure for the next block and store the progress

	}

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
