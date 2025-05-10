package structures

type ApprovementThread struct {
	CoreMajorVersion  int                     `json:"coreMajorVersion"`
	NetworkParameters NetworkParameters       `json:"networkParameters"`
	EpochHandler      EpochHandler            `json:"epoch"`
	Cache             map[string]*PoolStorage `json:"-"`
}

type GenerationThread struct {
	EpochFullId string   `json:"epochFullId"`
	EpochIndex  int      `json:"epochIndex"`
	PrevHash    string   `json:"prevHash"`
	NextIndex   int      `json:"nextIndex"`
	Quorum      []string `json:"quorum"`
	Majority    int      `json:"majority"`
}
