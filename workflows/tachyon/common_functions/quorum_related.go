package common_functions

import (
	"math"
	"math/big"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon/threads"
)

type ValidatorData struct {
	ValidatorPubKey string
	TotalStake      *big.Int
}

func setLeadersSequence(epochHandler *threads.EpochHandler, epochSeed string) error {

	// epochHandler.LeaderSequence = []string{} // [pool0, pool1,...poolN]

	// // Hash of metadata from the old epoch

	// hashOfMetadataFromOldEpoch := utils.Blake3(epochSeed)

	// // Change order of validators pseudo-randomly

	// validatorsExtendedData := make(map[string]ValidatorData)

	// var totalStakeSum *big.Int

	// totalStakeSum = big.NewInt(0)

	// // Populate validator data and calculate total stake sum

	// for _, validatorPubKey := range epochHandler.PoolsRegistry {

	// 	validatorData, err := getFromApprovementThreadState(fmt.Sprintf("%v(POOL)_STORAGE_POOL", validatorPubKey))

	// 	if err != nil {
	// 		return err
	// 	}

	// 	// Calculate total stake

	// 	totalStake := new(big.Int)

	// 	totalStake.Add(new(big.Int).SetInt64(validatorData.TotalStakedKly), new(big.Int).SetInt64(validatorData.TotalStakedUno))

	// 	totalStakeSum.Add(totalStakeSum, totalStake)

	// 	validatorsExtendedData[validatorPubKey] = ValidatorData{
	// 		ValidatorPubKey: validatorPubKey,
	// 		TotalStake:      totalStake,
	// 	}
	// }

	// // Iterate over the poolsRegistry and pseudo-randomly choose leaders
	// for i := 0; i < len(epochHandler.PoolsRegistry); i++ {
	// 	cumulativeSum := big.NewInt(0)

	// 	// Generate deterministic random value using the hash of metadata
	// 	hashInput := fmt.Sprintf("%v_%v", hashOfMetadataFromOldEpoch, i)
	// 	deterministicRandomValue := new(big.Int)
	// 	deterministicRandomValue.SetString(blake3Hash(hashInput), 16)
	// 	deterministicRandomValue.Mod(deterministicRandomValue, totalStakeSum)

	// 	// Find the validator based on the random value
	// 	for validatorPubKey, validator := range validatorsExtendedData {
	// 		cumulativeSum.Add(cumulativeSum, validator.TotalStake)

	// 		if deterministicRandomValue.Cmp(cumulativeSum) <= 0 {
	// 			// Add the chosen validator to the leaders sequence
	// 			epochHandler.LeaderSequence = append(epochHandler.LeaderSequence, validatorPubKey)

	// 			// Update totalStakeSum and remove the chosen validator from the map
	// 			totalStakeSum.Sub(totalStakeSum, validator.TotalStake)
	// 			delete(validatorsExtendedData, validatorPubKey)

	// 			break
	// 		}
	// 	}
	// }

	// return nil

	return nil
}

func getQuorumMajority(epochHandler *threads.EpochHandler) uint {

	quorumSize := len(epochHandler.Quorum)

	majority := int(math.Floor(float64(quorumSize)*2/3)) + 1

	if majority > quorumSize {

		return uint(quorumSize)

	}

	return uint(majority)

}
