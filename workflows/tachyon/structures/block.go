package structures

import (
	"encoding/json"

	klyUtils "github.com/KlyntarNetwork/KlyntarCoreGolang/utils"

	klyGlobals "github.com/KlyntarNetwork/KlyntarCoreGolang/workflows/tachyon"
)

type Block struct {
	Creator      string        `json:"creator"`
	Time         uint64        `json:"time"`
	Epoch        string        `json:"epoch"`
	Transactions []Transaction `json:"transactions"`
	ExtraData    []string      `json:"extraData"`
	Index        uint32        `json:"index"`
	PrevHash     string        `json:"prevHash"`
	Sig          string        `json:"sig"`
}

func (block *Block) getHash() string {

	jsonedTransactions, _ := json.Marshal(block.Transactions)

	networkID, _ := klyGlobals.GENESIS["NETWORK_ID"].(string)

	dataToHash := block.Creator + string(block.Time) + string(jsonedTransactions) + networkID + block.Epoch + string(block.Index) + block.PrevHash

	return klyUtils.Blake3(dataToHash)

}
