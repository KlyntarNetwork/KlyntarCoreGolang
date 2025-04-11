package structures

import (
	"encoding/json"

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

	networkID, _ := tachyon.GENESIS["NETWORK_ID"].(string)

	dataToHash := block.Creator + string(block.Time) + string(jsonedTransactions) + networkID + block.Epoch + string(block.Index) + block.PrevHash

	return utils.Blake3(dataToHash)

}
