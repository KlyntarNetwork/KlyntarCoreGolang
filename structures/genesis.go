package structures

import "math/big"

type NetworkParameters struct {
	ValidatorStake        int64 `json:"VALIDATOR_STAKE"`
	MinimalStakePerEntity int64 `json:"MINIMAL_STAKE_PER_ENTITY"`
	QuorumSize            int   `json:"QUORUM_SIZE"`
	EpochTime             int64 `json:"EPOCH_TIME"`
	LeadershipTimeframe   int64 `json:"LEADERSHIP_TIMEFRAME"`
	BlockTime             int64 `json:"BLOCK_TIME"`
	MaxBlockSizeInBytes   int64 `json:"MAX_BLOCK_SIZE_IN_BYTES"`
	TxLimitPerBlock       int   `json:"TXS_LIMIT_PER_BLOCK"`
}

type Staker struct {
	Kly *big.Int `json:"kly"`
	Uno *big.Int `json:"uno"`
}

type PoolStorage struct {
	Percentage     int               `json:"percentage"`
	TotalStakedKly *big.Int          `json:"totalStakedKly"`
	TotalStakedUno *big.Int          `json:"totalStakedUno"`
	Stakers        map[string]Staker `json:"stakers"`
	PoolURL        string            `json:"poolURL"`
	WssPoolURL     string            `json:"wssPoolURL"`
	Activated      bool              `json:"activated"`
}

type Genesis struct {
	NetworkID                string                 `json:"NETWORK_ID"`
	Shard                    string                 `json:"SHARD"`
	CoreMajorVersion         int                    `json:"CORE_MAJOR_VERSION"`
	FirstEpochStartTimestamp uint64                 `json:"FIRST_EPOCH_START_TIMESTAMP"`
	NetworkCreatorsContact   map[string]string      `json:"NETWORK_CREATORS_CONTACT"`
	HiveMind                 []string               `json:"HIVEMIND"`
	Hostchains               map[string]string      `json:"HOSTCHAINS"`
	NetworkWorkflow          string                 `json:"NETWORK_WORKFLOW"`
	NetworkParameters        NetworkParameters      `json:"NETWORK_PARAMETERS"`
	Pools                    map[string]PoolStorage `json:"POOLS"`
}
