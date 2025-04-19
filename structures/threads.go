package structures

type NetworkParams struct {
	ValidatorStake        int64 `json:"validatorStake"`
	MinimalStakePerEntity int64 `json:"minimalStakePerEntity"`
	QuorumSize            int   `json:"quorumSize"`
	EpochTime             int64 `json:"epochTime"`
	LeadershipTimeframe   int64 `json:"leadershipTimeframe"`
	BlockTime             int64 `json:"blockTime"`
	MaxBlockSizeInBytes   int64 `json:"maxBlockSizeInBytes"`
	TxsLimitPerBlock      int   `json:"txsLimitPerBlock"`
}

type ApprovementThread struct {
	CoreMajorVersion  uint          `json:"coreMajorVersion"`
	NetworkParameters NetworkParams `json:"networkParameters"`
	Epoch             EpochHandler  `json:"epoch"`
}

type GenerationThread struct {
	EpochFullId string   `json:"epochFullId"`
	EpochIndex  int      `json:"epochIndex"`
	PrevHash    string   `json:"prevHash"`
	NextIndex   int      `json:"nextIndex"`
	Quorum      []string `json:"quorum"`
	Majority    uint     `json:"majority"`
}
