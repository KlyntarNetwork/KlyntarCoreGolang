package system_contracts

import (
	"math/big"
)

type DelayedTransaction struct {
	Type    string
	Payload any
}

type CreateStakingPoolTx struct {
	Type       string `json:"type"`
	Creator    string `json:"creator"`
	Percentage int    `json:"percentage"`
	PoolURL    string `json:"poolURL"`
	WssPoolURL string `json:"wssPoolURL"`
}

type UpdateStakingPoolTx struct {
	Type       string `json:"type"`
	Creator    string `json:"creator"`
	Activated  bool   `json:"activated"`
	Percentage int    `json:"percentage"`
	PoolURL    string `json:"poolURL"`
	WssPoolURL string `json:"wssPoolURL"`
}

type StakeTx struct {
	Type       string   `json:"type"`
	Staker     string   `json:"staker"`
	PoolPubKey string   `json:"poolPubKey"`
	Amount     *big.Int `json:"amount"`
}

type UnstakeTx struct {
	Type       string   `json:"type"`
	Unstaker   string   `json:"unstaker"`
	PoolPubKey string   `json:"poolPubKey"`
	Amount     *big.Int `json:"amount"`
}

type ChangeUnobtaniumAmountTx struct {
	Type               string              `json:"type"`
	TargetPool         string              `json:"targetPool"`
	ChangesPerAccounts map[string]*big.Int `json:"changesPerAccounts"`
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
}
