package structures

type EpochFinishProposition struct {
	CurrentLeader        uint                        `json:"currentLeader"`
	AfpForFirstBlock     AggregatedFinalizationProof `json:"afpForFirstBlock"`
	LastBlockProposition PoolVotingStat              `json:"lastBlockProposition"`
}
