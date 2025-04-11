package threads

type EpochHandler struct {
	Id                                                    uint
	Hash                                                  string
	PoolsRegistry, ShardsRegistry, Quorum, LeaderSequence []string
	StartTimestamp                                        uint64
}
