package threads

type NetworkParams struct {
	ValidatorStake        int64 `json:"VALIDATOR_STAKE"`
	MinimalStakePerEntity int64 `json:"MINIMAL_STAKE_PER_ENTITY"`
	QuorumSize            int   `json:"QUORUM_SIZE"`
	EpochTime             int64 `json:"EPOCH_TIME"`
	LeadershipTimeframe   int64 `json:"LEADERSHIP_TIMEFRAME"`
	BlockTime             int64 `json:"BLOCK_TIME"`
	MaxBlockSizeInBytes   int64 `json:"MAX_BLOCK_SIZE_IN_BYTES"`
	TxsLimitPerBlock      int   `json:"TXS_LIMIT_PER_BLOCK"`
}

type ApprovementThread struct {
	CoreMajorVersion  uint
	NetworkParameters NetworkParams
	Epoch             EpochHandler
}
