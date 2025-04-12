package threads

type EpochHandler struct {
	Id             uint     `json:"id"`
	Hash           string   `json:"hash"`
	PoolsRegistry  []string `json:"poolsRegistry"`
	ShardsRegistry []string `json:"shardsRegistry"`
	Quorum         []string `json:"quorum"`
	LeaderSequence []string `json:"leaderSequence"`
	StartTimestamp uint64   `json:"startTimestamp"`
}
