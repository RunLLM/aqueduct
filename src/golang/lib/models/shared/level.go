package shared

type Level string

const (
	SuccessLevel Level = "success"
	WarningLevel Level = "warning"
	ErrorLevel   Level = "error"
	InfoLevel    Level = "info"
	NeutralLevel Level = "neutral"
)