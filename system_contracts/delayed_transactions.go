package system_contracts

import (
	"math/big"
	"strconv"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/common_functions"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
)

type DelayedTxExecutorFunction = func(map[string]string) bool

var DELAYED_TRANSACTIONS_MAP = map[string]DelayedTxExecutorFunction{
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

		if _, exists := globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.Cache[storageKey]; exists {

			return false

		}

		globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.Cache[storageKey] = &structures.PoolStorage{
			Percentage:     percentage,
			TotalStakedKly: structures.BigInt{Int: big.NewInt(0)},
			TotalStakedUno: structures.BigInt{Int: big.NewInt(0)},
			Stakers: map[string]structures.Staker{
				creator: {
					Kly: structures.BigInt{Int: big.NewInt(0)},
					Uno: structures.BigInt{Int: big.NewInt(0)},
				},
			},
			PoolUrl:    poolURL,
			WssPoolUrl: wssPoolURL,
		}

		return true
	}

	return false
}

func UpdateStakingPool(delayedTransaction map[string]string) bool {

	creator := delayedTransaction["creator"]
	percentage, err1 := strconv.Atoi(delayedTransaction["percentage"])
	poolURL := delayedTransaction["poolURL"]
	wssPoolURL := delayedTransaction["wssPoolURL"]

	if err1 != nil || percentage < 0 || percentage > 100 || poolURL == "" || wssPoolURL == "" {

		return false

	}

	poolStorage := common_functions.GetFromApprovementThreadState(creator + "(POOL)_STORAGE_POOL")

	if poolStorage != nil {

		poolStorage.Percentage = percentage
		poolStorage.PoolUrl = poolURL
		poolStorage.WssPoolUrl = wssPoolURL

		requiredStake := globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.NetworkParameters.ValidatorStake

		if poolStorage.TotalStakedKly.Int.Cmp(requiredStake.Int) >= 0 {
			globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.EpochDataHandler.PoolsRegistry[creator] = struct{}{}
		} else {
			delete(globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.EpochDataHandler.PoolsRegistry, creator)
		}

		globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.Cache[creator+"(POOL)_STORAGE_POOL"] = poolStorage

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

		minStake := globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.NetworkParameters.MinimalStakePerEntity

		if amount.Cmp(minStake.Int) < 0 {

			return false

		}

		if _, exists := poolStorage.Stakers[staker]; !exists {

			poolStorage.Stakers[staker] = structures.Staker{

				Kly: structures.BigInt{Int: big.NewInt(0)},
				Uno: structures.BigInt{Int: big.NewInt(0)},
			}

		}

		stakerData := poolStorage.Stakers[staker]
		stakerData.Kly = structures.BigInt{Int: new(big.Int).Add(stakerData.Kly.Int, amount)}
		poolStorage.TotalStakedKly = structures.BigInt{Int: new(big.Int).Add(poolStorage.TotalStakedKly.Int, amount)}
		poolStorage.Stakers[staker] = stakerData

		requiredStake := globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.NetworkParameters.ValidatorStake

		if poolStorage.TotalStakedKly.Cmp(requiredStake.Int) >= 0 {

			if _, exists := globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.EpochDataHandler.PoolsRegistry[poolPubKey]; !exists {

				globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.EpochDataHandler.PoolsRegistry[poolPubKey] = struct{}{}

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

		stakerData.Kly.Sub(stakerData.Kly.Int, amount)

		poolStorage.TotalStakedKly.Sub(poolStorage.TotalStakedKly.Int, amount)

		if stakerData.Kly.Cmp(big.NewInt(0)) == 0 && stakerData.Uno.Cmp(big.NewInt(0)) == 0 {

			delete(poolStorage.Stakers, unstaker)

		} else {

			poolStorage.Stakers[unstaker] = stakerData

		}

		requiredStake := globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.NetworkParameters.ValidatorStake

		if poolStorage.TotalStakedKly.Cmp(requiredStake.Int) < 0 {

			delete(globals.APPROVEMENT_THREAD_METADATA_HANDLER.Handler.EpochDataHandler.PoolsRegistry, poolPubKey)

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
				stakerData = structures.Staker{Kly: structures.BigInt{Int: big.NewInt(0)}, Uno: structures.BigInt{Int: big.NewInt(0)}}
			}

			stakerData.Uno = structures.BigInt{Int: new(big.Int).Add(stakerData.Uno.Int, delta)}

			if stakerData.Uno.Sign() < 0 {

				stakerData.Uno = structures.BigInt{Int: big.NewInt(0)}

			}

			if stakerData.Kly.Sign() == 0 && stakerData.Uno.Sign() == 0 {

				delete(poolStorage.Stakers, key)

			} else {

				poolStorage.Stakers[key] = stakerData

			}

			totalChange = new(big.Int).Add(totalChange, delta)
		}

		poolStorage.TotalStakedUno = structures.BigInt{Int: new(big.Int).Add(poolStorage.TotalStakedUno.Int, totalChange)}

		return true

	}

	return false

}
