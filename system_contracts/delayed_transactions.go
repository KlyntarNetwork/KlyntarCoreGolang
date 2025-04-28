package system_contracts

import (
	"math/big"
	"strconv"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
)

type DelayedTransactionsBatch struct {
	EpochIndex          int                 `json:"epochIndex"`
	DelayedTransactions []map[string]string `json:"delayedTransactions"`
	Proofs              map[string]string   `json:"proofs"`
}

type DelayedTxHandler = func(map[string]string) bool

var DELAYED_TRANSACTIONS_MAP = map[string]DelayedTxHandler{
	"createStakingPool":      CreateStakingPool,
	"updateStakingPool":      UpdateStakingPool,
	"stake":                  Stake,
	"unstake":                Unstake,
	"changeUnobtaniumAmount": ChangeUnobtaniumAmount,
}

func CreateStakingPool(delayedTransaction map[string]string) bool {

	creator := delayedTransaction["creator"]
	percentage, _ := strconv.Atoi(delayedTransaction["percentage"])
	poolURL := delayedTransaction["poolURL"]
	wssPoolURL := delayedTransaction["wssPoolURL"]

	if poolURL != "" && wssPoolURL != "" && percentage >= 0 && percentage <= 100 {

		storageKey := creator + "(POOL)_STORAGE_POOL"

		if _, exists := globals.APPROVEMENT_THREAD.Cache[storageKey]; exists {

			return false

		}

		globals.APPROVEMENT_THREAD.Cache[storageKey] = &structures.PoolStorage{
			Activated:      true,
			Percentage:     percentage,
			TotalStakedKly: big.NewInt(0),
			TotalStakedUno: big.NewInt(0),
			Stakers: map[string]structures.Staker{
				creator: {
					Kly: big.NewInt(0),
					Uno: big.NewInt(0),
				},
			},
			PoolURL:    poolURL,
			WssPoolURL: wssPoolURL,
		}

		return true
	}

	return false
}

func UpdateStakingPool(delayedTransaction map[string]string) bool {

	creator := delayedTransaction["creator"]
	percentage, err1 := strconv.Atoi(delayedTransaction["percentage"])
	activated, err2 := strconv.ParseBool(delayedTransaction["activated"])
	poolURL := delayedTransaction["poolURL"]
	wssPoolURL := delayedTransaction["wssPoolURL"]

	if err1 != nil || err2 != nil || percentage < 0 || percentage > 100 || poolURL == "" || wssPoolURL == "" {

		return false

	}

	poolStorage := common_functions.GetFromApprovementThreadState(creator + "(POOL)_STORAGE_POOL")

	if poolStorage != nil {

		poolStorage.Activated = activated
		poolStorage.Percentage = percentage
		poolStorage.PoolURL = poolURL
		poolStorage.WssPoolURL = wssPoolURL

		requiredStake := globals.APPROVEMENT_THREAD.NetworkParameters.ValidatorStake

		if activated {
			if poolStorage.TotalStakedKly.Cmp(requiredStake) >= 0 {
				globals.APPROVEMENT_THREAD.Epoch.PoolsRegistry[creator] = struct{}{}
			}
		} else {
			delete(globals.APPROVEMENT_THREAD.Epoch.PoolsRegistry, creator)
		}

		globals.APPROVEMENT_THREAD.Cache[creator+"(POOL)_STORAGE_POOL"] = poolStorage

		return true

	}

	return false

}

func Stake(delayedTransaction map[string]string) bool {

	staker := delayedTransaction["staker"]
	poolPubKey := delayedTransaction["poolPubKey"]
	amount, ok := new(big.Int).SetString(delayedTransaction["amount"], 10)

	if !ok {

		return false

	}

	poolStorage := common_functions.GetFromApprovementThreadState(poolPubKey + "(POOL)_STORAGE_POOL")

	if poolStorage != nil {

		minStake := globals.APPROVEMENT_THREAD.NetworkParameters.MinimalStakePerEntity

		if amount.Cmp(minStake) < 0 {

			return false

		}

		if _, exists := poolStorage.Stakers[staker]; !exists {

			poolStorage.Stakers[staker] = structures.Staker{Kly: big.NewInt(0), Uno: big.NewInt(0)}

		}

		stakerData := poolStorage.Stakers[staker]
		stakerData.Kly = new(big.Int).Add(stakerData.Kly, amount)
		poolStorage.TotalStakedKly = new(big.Int).Add(poolStorage.TotalStakedKly, amount)
		poolStorage.Stakers[staker] = stakerData

		requiredStake := globals.APPROVEMENT_THREAD.NetworkParameters.ValidatorStake

		if poolStorage.Activated && poolStorage.TotalStakedKly.Cmp(requiredStake) >= 0 {

			if _, exists := globals.APPROVEMENT_THREAD.Epoch.PoolsRegistry[poolPubKey]; !exists {

				globals.APPROVEMENT_THREAD.Epoch.PoolsRegistry[poolPubKey] = struct{}{}

			}

		}

		return true

	}

	return false

}

func Unstake(delayedTransaction map[string]string) bool {

	unstaker := delayedTransaction["unstaker"]
	poolPubKey := delayedTransaction["poolPubKey"]
	amount, ok := new(big.Int).SetString(delayedTransaction["amount"], 10)

	if !ok {

		return false

	}

	poolStorage := common_functions.GetFromApprovementThreadState(poolPubKey + "(POOL)_STORAGE_POOL")

	if poolStorage != nil {

		stakerData, exists := poolStorage.Stakers[unstaker]

		if !exists {

			return false

		}

		if stakerData.Kly.Cmp(amount) < 0 {

			return false

		}

		stakerData.Kly.Sub(stakerData.Kly, amount)

		poolStorage.TotalStakedKly.Sub(poolStorage.TotalStakedKly, amount)

		if stakerData.Kly.Cmp(big.NewInt(0)) == 0 && stakerData.Uno.Cmp(big.NewInt(0)) == 0 {

			delete(poolStorage.Stakers, unstaker)

		} else {

			poolStorage.Stakers[unstaker] = stakerData

		}

		requiredStake := globals.APPROVEMENT_THREAD.NetworkParameters.ValidatorStake

		if poolStorage.TotalStakedKly.Cmp(requiredStake) < 0 {

			delete(globals.APPROVEMENT_THREAD.Epoch.PoolsRegistry, poolPubKey)

		}

		return true

	}

	return false

}

func ChangeUnobtaniumAmount(delayedTransaction map[string]string) bool {

	targetPool := delayedTransaction["targetPool"]

	poolStorage := common_functions.GetFromApprovementThreadState(targetPool + "(POOL)_STORAGE_POOL")

	if poolStorage != nil {

		totalChange := big.NewInt(0)

		for key, unoDeltaStr := range delayedTransaction {

			if key == "type" || key == "targetPool" {
				continue
			}

			delta, ok := new(big.Int).SetString(unoDeltaStr, 10)

			if !ok {
				return false
			}

			stakerData, exists := poolStorage.Stakers[key]

			if !exists {
				stakerData = structures.Staker{Kly: big.NewInt(0), Uno: big.NewInt(0)}
			}

			stakerData.Uno = new(big.Int).Add(stakerData.Uno, delta)

			if stakerData.Uno.Sign() < 0 {

				stakerData.Uno = big.NewInt(0)

			}

			if stakerData.Kly.Sign() == 0 && stakerData.Uno.Sign() == 0 {

				delete(poolStorage.Stakers, key)

			} else {

				poolStorage.Stakers[key] = stakerData

			}

			totalChange = new(big.Int).Add(totalChange, delta)
		}

		poolStorage.TotalStakedUno = new(big.Int).Add(poolStorage.TotalStakedUno, totalChange)

		return true

	}

	return false

}
