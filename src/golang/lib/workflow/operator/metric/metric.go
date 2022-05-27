package metric

import (
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/function"
)

type Metric struct {
	Function function.Function `json:"function"`
}
