package structures

type ApprovementThread struct {
	CoreMajorVersion  int                     `json:"coreMajorVersion"`
	NetworkParameters NetworkParameters       `json:"networkParameters"`
	EpochHandler      EpochHandler            `json:"epoch"`
	Cache             map[string]*PoolStorage `json:"-"`
}

type GenerationThread struct {
	EpochFullId string `json:"epochFullId"`
	PrevHash    string `json:"prevHash"`
	NextIndex   int    `json:"nextIndex"`
}
