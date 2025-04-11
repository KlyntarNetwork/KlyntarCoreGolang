package common_functions

import (
	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/threads"
)

func VerifyAggregatedEpochFinalizationProof(proofStruct *structures.AggregatedEpochFinalizationProof) {
}

func VerifyAggregatedFinalizationProof(proofStruct *structures.AggregatedFinalizationProof) {
}

func GetVerifiedAggregatedFinalizationProofByBlockId(blockID string, epochHandler *threads.EpochHandler) {
}

func GetFirstBlockInEpoch(epochHandler *threads.EpochHandler) {}

func CheckAggregatedLeaderRotationProofValidity() {}

func CheckAlrpChainValidity() {}
