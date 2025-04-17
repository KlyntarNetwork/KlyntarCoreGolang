package structures

type NetworkCreatorsContact struct {
	Telegram string `json:"telegram"`
	Email    string `json:"email"`
	Tor      string `json:"tor"`
}

type Hostchain struct {
	BlockWithGenesisCommit int `json:"blockWithGenesisCommit"`
}

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
	Kly string `json:"kly"`
	Uno string `json:"uno"`
}

type Pool struct {
	Percentage     int               `json:"percentage"`
	TotalStakedKly string            `json:"totalStakedKly"`
	TotalStakedUno string            `json:"totalStakedUno"`
	Stakers        map[string]Staker `json:"stakers"`
	PoolURL        string            `json:"poolURL"`
	WssPoolURL     string            `json:"wssPoolURL"`
}

type Genesis struct {
	NetworkID                string                 `json:"NETWORK_ID"`
	Shard                    string                 `json:"SHARD"`
	CoreMajorVersion         int                    `json:"CORE_MAJOR_VERSION"`
	FirstEpochStartTimestamp int64                  `json:"FIRST_EPOCH_START_TIMESTAMP"`
	NetworkCreatorsContact   NetworkCreatorsContact `json:"NETWORK_CREATORS_CONTACT"`
	HiveMind                 []string               `json:"HIVEMIND"`
	Hostchains               map[string]Hostchain   `json:"HOSTCHAINS"`
	NetworkWorkflow          string                 `json:"NETWORK_WORKFLOW"`
	NetworkParameters        NetworkParameters      `json:"NETWORK_PARAMETERS"`
	Pools                    map[string]Pool        `json:"POOLS"`
}
