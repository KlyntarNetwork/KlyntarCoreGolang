package structures

type ApprovementThreadMetadataHandler struct {
	CoreMajorVersion  int                     `json:"coreMajorVersion"`
	NetworkParameters NetworkParameters       `json:"networkParameters"`
	EpochDataHandler  EpochDataHandler        `json:"epoch"`
	Cache             map[string]*PoolStorage `json:"-"`
}

type GenerationThreadMetadataHandler struct {
	EpochFullId string `json:"epochFullId"`
	PrevHash    string `json:"prevHash"`
	NextIndex   int    `json:"nextIndex"`
}
