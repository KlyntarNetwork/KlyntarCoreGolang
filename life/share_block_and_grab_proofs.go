package life

import (
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/gorilla/websocket"
)

var WEBSOCKET_CONNECTIONS map[string]*websocket.Conn

var FINALIZATION_PROOFS_CACHE map[string]string

var RESPONSES chan Agreement

var PROOFS_GRABBER struct {
	EpochId        string
	AcceptedIndex  int
	AcceptedHash   string
	AfpForPrevious structures.AggregatedFinalizationProof
}

func processIncomingFinalizationProof(msg []byte) {}

func BlocksSharingAndProofsGrabingThread() {}
