package tachyon_structures

import (
	"encoding/json"

	klyUtils "github.com/KLYN74R/KlyntarCoreGolang/KLY_Utils"

	klyGlobals "github.com/KLYN74R/KlyntarCoreGolang/KLY_Globals"
)

type Block struct {
	creator string

	time uint64

	epoch string

	transactions []Transaction

	extraData []string

	index uint32

	prevHash string

	sig string
}

func (block *Block) getHash() string {

	jsonedTransactions, _ := json.Marshal(block.transactions)

	symbioticChainID, _ := klyGlobals.GENESIS["SYMBIOTE_ID"].(string)

	dataToHash := block.creator + string(block.time) + string(jsonedTransactions) + symbioticChainID + block.epoch + string(block.index) + block.prevHash

	return klyUtils.Blake3(dataToHash)

}
