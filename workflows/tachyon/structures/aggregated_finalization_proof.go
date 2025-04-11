package structures

type AggregatedFinalizationProof struct {
	PrevBlockHash string   `json:"prevBlockHash"`
	BlockID       string   `json:"blockId"`
	BlockHash     string   `json:"blockHash"`
	Proofs        []string `json:"proofs"`
}
