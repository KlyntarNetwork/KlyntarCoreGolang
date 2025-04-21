package life

import "github.com/KlyntarNetwork/KlyntarCoreGolang/structures"

type FirstBlockDataWithAefp struct {
	FirstBlockCreator, FirstBlockHash string

	Aefp structures.AggregatedEpochFinalizationProof
}

var AEFP_AND_FIRST_BLOCK_DATA FirstBlockDataWithAefp

func EpochRotationThread() {}
