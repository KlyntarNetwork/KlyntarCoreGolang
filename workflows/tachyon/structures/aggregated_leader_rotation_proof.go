package structures

type AggregatedLeaderRotationProof struct {
	FirstBlockHash string            `json:"firstBlockHash"`
	SkipIndex      uint              `json:"skipIndex"`
	SkipHash       string            `json:"skipHash"`
	Proofs         map[string]string `json:"proofs"`
}
