package system_contracts

func CreateStakingPool(rawTx map[string]string) {}

func UpdateStakingPool(rawTx map[string]string) {}

func Stake(rawTx map[string]string) {}

func Unstake(rawTx map[string]string) {}

func ChangeUnobtaniumAmount(rawTx map[string]string) {}

var DELAYED_TRANSACTIONS_MAP = map[string]func(map[string]string){
	"createStakingPool":      CreateStakingPool,
	"updateStakingPool":      UpdateStakingPool,
	"stake":                  Stake,
	"unstake":                Unstake,
	"changeUnobtaniumAmount": ChangeUnobtaniumAmount,
}
