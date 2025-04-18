package common_functions

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
)

type ValidatorData struct {
	ValidatorPubKey string
	TotalStake      *big.Int
}

func GetFromApprovementThreadState(poolId string) *structures.Pool {

	if val, ok := globals.APPROVEMENT_THREAD_CACHE[poolId]; ok {
		return val
	}

	data, err := globals.APPROVEMENT_THREAD_METADATA.Get([]byte(poolId), nil)
	if err != nil {
		return nil
	}

	var pool structures.Pool

	err = json.Unmarshal(data, &pool)

	if err != nil {
		return nil
	}

	globals.APPROVEMENT_THREAD_CACHE[poolId] = &pool

	return &pool

}

func SetLeadersSequence(epochHandler *structures.EpochHandler, epochSeed string) {

	epochHandler.LeaderSequence = []string{} // [pool0, pool1,...poolN]

	// Hash of metadata from the old epoch

	hashOfMetadataFromOldEpoch := utils.Blake3(epochSeed)

	// Change order of validators pseudo-randomly

	validatorsExtendedData := make(map[string]ValidatorData)

	var totalStakeSum *big.Int = big.NewInt(0)

	// Populate validator data and calculate total stake sum

	for _, validatorPubKey := range epochHandler.PoolsRegistry {

		validatorData := GetFromApprovementThreadState(fmt.Sprintf("%v(POOL)_STORAGE_POOL", validatorPubKey))

		// Calculate total stake

		totalStakeByThisValidator := new(big.Int)

		totalStakeByThisValidator.Add(totalStakeByThisValidator, validatorData.TotalStakedKly)
		totalStakeByThisValidator.Add(totalStakeByThisValidator, validatorData.TotalStakedUno)

		totalStakeSum.Add(totalStakeSum, totalStakeByThisValidator)

		validatorsExtendedData[validatorPubKey] = ValidatorData{validatorPubKey, totalStakeByThisValidator}

	}

	// Iterate over the poolsRegistry and pseudo-randomly choose leaders

	for i := 0; i < len(epochHandler.PoolsRegistry); i++ {

		cumulativeSum := big.NewInt(0)

		// Generate deterministic random value using the hash of metadata
		hashInput := fmt.Sprintf("%v_%v", hashOfMetadataFromOldEpoch, i)
		deterministicRandomValue := new(big.Int)
		deterministicRandomValue.SetString(utils.Blake3(hashInput), 16)
		deterministicRandomValue.Mod(deterministicRandomValue, totalStakeSum)

		// Find the validator based on the random value
		for validatorPubKey, validator := range validatorsExtendedData {

			cumulativeSum.Add(cumulativeSum, validator.TotalStake)

			if deterministicRandomValue.Cmp(cumulativeSum) <= 0 {

				// Add the chosen validator to the leaders sequence
				epochHandler.LeaderSequence = append(epochHandler.LeaderSequence, validatorPubKey)

				// Update totalStakeSum and remove the chosen validator from the map
				totalStakeSum.Sub(totalStakeSum, validator.TotalStake)
				delete(validatorsExtendedData, validatorPubKey)

				break

			}

		}

	}

}

func GetQuorumMajority(epochHandler *structures.EpochHandler) uint {

	quorumSize := len(epochHandler.Quorum)

	majority := (2 * quorumSize) / 3

	majority += 1

	if majority > quorumSize {
		return uint(quorumSize)
	}

	return uint(majority)
}

func GetQuorumUrlsAndPubkeys(epochHandler *structures.EpochHandler) []structures.QuorumMemberData {

	var toReturn []structures.QuorumMemberData

	for _, pubKey := range epochHandler.Quorum {

		poolStorage := GetFromApprovementThreadState(pubKey + "(POOL)_STORAGE_POOL")

		toReturn = append(toReturn, structures.QuorumMemberData{PubKey: pubKey, Url: poolStorage.PoolURL})

	}

	return toReturn

}

func GetCurrentEpochQuorum(epochHandler *structures.EpochHandler, quorumSize int, newEpochSeed string) []string {

	if len(epochHandler.PoolsRegistry) <= quorumSize {

		return epochHandler.PoolsRegistry

	}

	quorum := []string{}

	hashOfMetadataFromEpoch := utils.Blake3(newEpochSeed)

	validatorsExtendedData := make(map[string]ValidatorData)
	totalStakeSum := big.NewInt(0)

	for _, validatorPubKey := range epochHandler.PoolsRegistry {
		validatorData := GetFromApprovementThreadState(fmt.Sprintf("%v(POOL)_STORAGE_POOL", validatorPubKey))

		totalStakeByThisValidator := new(big.Int)
		totalStakeByThisValidator.Add(totalStakeByThisValidator, validatorData.TotalStakedKly)
		totalStakeByThisValidator.Add(totalStakeByThisValidator, validatorData.TotalStakedUno)

		validatorsExtendedData[validatorPubKey] = ValidatorData{
			ValidatorPubKey: validatorPubKey,
			TotalStake:      totalStakeByThisValidator,
		}

		totalStakeSum.Add(totalStakeSum, totalStakeByThisValidator)
	}

	for i := range quorumSize {

		cumulativeSum := big.NewInt(0)

		hashInput := fmt.Sprintf("%v_%v", hashOfMetadataFromEpoch, i)
		deterministicRandomValue := new(big.Int)
		deterministicRandomValue.SetString(utils.Blake3(hashInput), 16)
		deterministicRandomValue.Mod(deterministicRandomValue, totalStakeSum)

		for validatorPubKey, validator := range validatorsExtendedData {

			cumulativeSum.Add(cumulativeSum, validator.TotalStake)

			if deterministicRandomValue.Cmp(cumulativeSum) <= 0 {

				quorum = append(quorum, validatorPubKey)

				totalStakeSum.Sub(totalStakeSum, validator.TotalStake)

				delete(validatorsExtendedData, validatorPubKey)

				break

			}

		}

	}

	return quorum

}
