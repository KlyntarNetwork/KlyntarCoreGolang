package structures

type EpochHandler struct {
	Id                 uint     `json:"id"`
	Hash               string   `json:"hash"`
	PoolsRegistry      []string `json:"poolsRegistry"`
	ShardsRegistry     []string `json:"shardsRegistry"`
	Quorum             []string `json:"quorum"`
	LeadersSequence    []string `json:"leadersSequence"`
	StartTimestamp     uint64   `json:"startTimestamp"`
	CurrentLeaderIndex uint     `json:"currentLeaderIndex"`
}
