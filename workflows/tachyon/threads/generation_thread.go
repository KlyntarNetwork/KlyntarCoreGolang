package threads

type GenerationThread struct {
	EpochFullId string `json:"epochFullId"`
	EpochIndex  int    `json:"epochIndex"`
	PrevHash    string `json:"prevHash"`
	NextIndex   int    `json:"nextIndex"`
}
