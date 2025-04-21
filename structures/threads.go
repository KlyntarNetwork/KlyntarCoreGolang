package structures

type ApprovementThread struct {
	CoreMajorVersion  int               `json:"coreMajorVersion"`
	NetworkParameters NetworkParameters `json:"networkParameters"`
	Epoch             EpochHandler      `json:"epoch"`
}

type GenerationThread struct {
	EpochFullId string   `json:"epochFullId"`
	EpochIndex  int      `json:"epochIndex"`
	PrevHash    string   `json:"prevHash"`
	NextIndex   int      `json:"nextIndex"`
	Quorum      []string `json:"quorum"`
	Majority    uint     `json:"majority"`
}
