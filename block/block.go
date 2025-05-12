package block

import (
	"encoding/json"
	"strconv"

	"github.com/KlyntarNetwork/KlyntarCoreGolang/globals"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/structures"
	"github.com/KlyntarNetwork/KlyntarCoreGolang/utils"
	"github.com/KlyntarNetwork/Web1337Golang/crypto_primitives/ed25519"
)

type Block struct {
	Creator      string                   `json:"creator"`
	Time         uint64                   `json:"time"`
	Epoch        string                   `json:"epoch"`
	Transactions []structures.Transaction `json:"transactions"`
	ExtraData    map[string]any           `json:"extraData"`
	Index        int                      `json:"index"`
	PrevHash     string                   `json:"prevHash"`
	Sig          string                   `json:"sig"`
}

func (block *Block) GetHash() string {

	jsonedTransactions, _ := json.Marshal(block.Transactions)

	networkID := globals.GENESIS.NetworkID

	dataToHash := block.Creator + strconv.FormatUint(block.Time, 10) + string(jsonedTransactions) + networkID + block.Epoch + strconv.FormatUint(uint64(block.Index), 10) + block.PrevHash

	return utils.Blake3(dataToHash)

}

func (block *Block) SignBlock() {

	block.Sig = ed25519.GenerateSignature(globals.CONFIGURATION.PrivateKey, block.GetHash())

}

func (block *Block) VerifySignature() bool {

	return ed25519.VerifySignature(block.GetHash(), block.Creator, block.Sig)

}
