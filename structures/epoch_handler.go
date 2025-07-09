package structures

type EpochDataHandler struct {
	Id                 int                 `json:"id"`
	Hash               string              `json:"hash"`
	PoolsRegistry      map[string]struct{} `json:"poolsRegistry"`
	ShardsRegistry     []string            `json:"shardsRegistry"`
	Quorum             []string            `json:"quorum"`
	LeadersSequence    []string            `json:"leadersSequence"`
	StartTimestamp     uint64              `json:"startTimestamp"`
	CurrentLeaderIndex int                 `json:"currentLeaderIndex"`
}
