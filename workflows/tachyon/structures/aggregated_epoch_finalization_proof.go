package structures

type AggregatedEpochFinalizationProof struct {
	LastLeader                   uint     `json:"lastLeader"`
	LastIndex                    uint     `json:"lastIndex"`
	LastHash                     string   `json:"lastHash"`
	HashOfFirstBlockByLastLeader string   `json:"hashOfFirstBlockByLastLeader"`
	Proofs                       []string `json:"proofs"`
}
