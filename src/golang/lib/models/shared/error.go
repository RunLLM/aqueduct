package shared

type Error struct {
	Context string `json:"context"`
	Tip     string `json:"tip"`
}
