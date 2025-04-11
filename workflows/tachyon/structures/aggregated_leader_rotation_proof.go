package structures

type AggregatedLeaderRotationProof struct {
	FirstBlockHash string   `json:"firstBlockHash"`
	SkipIndex      uint     `json:"skipIndex"`
	SkipHash       string   `json:"skipHash"`
	Proofs         []string `json:"proofs"`
}
