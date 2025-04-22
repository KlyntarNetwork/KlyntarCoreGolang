package system_contracts

func CreateStakingPool(rawTx []byte) {}

func UpdateStakingPool(rawTx []byte) {}

func Stake(rawTx []byte) {}

func Unstake(rawTx []byte) {}

func ChangeUnobtaniumAmount(rawTx []byte) {}

var DELAYED_TRANSACTIONS_MAP = map[string]func([]byte){
	"createStakingPool":      CreateStakingPool,
	"updateStakingPool":      UpdateStakingPool,
	"stake":                  Stake,
	"unstake":                Unstake,
	"changeUnobtaniumAmount": ChangeUnobtaniumAmount,
}
