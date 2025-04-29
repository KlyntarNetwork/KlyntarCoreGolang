package life

import (
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/lxzan/gws"
)

var WEBSOCKET_CONNECTIONS map[string]*gws.Conn

var FINALIZATION_PROOFS_CACHE map[string]string

var RESPONSES chan Agreement

var PROOFS_GRABBER struct {
	EpochId        string
	AcceptedIndex  int
	AcceptedHash   string
	AfpForPrevious structures.AggregatedFinalizationProof
}

func BlocksSharingAndProofsGrabingThread() {}
