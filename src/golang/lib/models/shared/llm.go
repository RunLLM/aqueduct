package shared

// LLMName specifies the name of the supported LLM.
type LLMName string

// Supported LLMs
const (
	Vicuna7b  LLMName = "vicuna_7b"
	DollyV23b LLMName = "dolly_v2_3b"
	DollyV27b LLMName = "dolly_v2_7b"
	Llama7b   LLMName = "llama_7b"
)
