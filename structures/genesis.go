package structures

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type BigInt struct {
	*big.Int
}

func (b *BigInt) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		i := new(big.Int)
		i, ok := i.SetString(s, 10)
		if !ok {
			return fmt.Errorf("invalid bigint string: %s", s)
		}
		b.Int = i
		return nil
	}

	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}

	i := new(big.Int)
	i, ok := i.SetString(num.String(), 10)
	if !ok {
		return fmt.Errorf("invalid bigint number: %s", num)
	}
	b.Int = i
	return nil
}

func (b BigInt) MarshalJSON() ([]byte, error) {
	if b.Int == nil {
		return []byte(`"0"`), nil
	}
	return json.Marshal(b.String())
}

type NetworkParameters struct {
	ValidatorStake        BigInt `json:"VALIDATOR_STAKE"`
	MinimalStakePerEntity BigInt `json:"MINIMAL_STAKE_PER_ENTITY"`
	QuorumSize            int    `json:"QUORUM_SIZE"`
	EpochTime             int64  `json:"EPOCH_TIME"`
	LeadershipTimeframe   int64  `json:"LEADERSHIP_TIMEFRAME"`
	BlockTime             int64  `json:"BLOCK_TIME"`
	MaxBlockSizeInBytes   int64  `json:"MAX_BLOCK_SIZE_IN_BYTES"`
	TxLimitPerBlock       int    `json:"TXS_LIMIT_PER_BLOCK"`
}

type Staker struct {
	Kly BigInt `json:"kly"`
	Uno BigInt `json:"uno"`
}

type PoolStorage struct {
	Percentage     int               `json:"percentage"`
	TotalStakedKly BigInt            `json:"totalStakedKly"`
	TotalStakedUno BigInt            `json:"totalStakedUno"`
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
