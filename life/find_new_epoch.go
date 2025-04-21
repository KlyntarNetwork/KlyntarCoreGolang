package life

import "github.com/KlyntarNetwork/KlyntarCoreGolang/structures"

type FirstBlockDataWithAefp struct {
	FirstBlockCreator, FirstBlockHash string

	Aefp structures.AggregatedEpochFinalizationProof
}

var aefpAndFirstBlockData FirstBlockDataWithAefp

func EpochRotationThread() {}
