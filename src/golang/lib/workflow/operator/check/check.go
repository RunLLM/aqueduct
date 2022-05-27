package check

import (
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/function"
)

type Level string

const (
	ErrorLevel   Level = "error"
	WarningLevel Level = "warning"
)

type Check struct {
	Level    Level             `json:"level"`
	Function function.Function `json:"function"`
}
