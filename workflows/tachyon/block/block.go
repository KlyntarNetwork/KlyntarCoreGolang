package block

import (
	"encoding/json"
	"strconv"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon"
)

type Block struct {
	Creator      string                `json:"creator"`
	Time         uint64                `json:"time"`
	Epoch        string                `json:"epoch"`
	Transactions []tachyon.Transaction `json:"transactions"`
	ExtraData    []string              `json:"extraData"`
	Index        uint32                `json:"index"`
	PrevHash     string                `json:"prevHash"`
	Sig          string                `json:"sig"`
}

func (block *Block) getHash() string {

	jsonedTransactions, _ := json.Marshal(block.Transactions)

	networkID := tachyon.GENESIS.NetworkID

	dataToHash := block.Creator + strconv.FormatUint(block.Time, 10) + string(jsonedTransactions) + networkID + block.Epoch + strconv.FormatUint(uint64(block.Index), 10) + block.PrevHash

	return utils.Blake3(dataToHash)

}
