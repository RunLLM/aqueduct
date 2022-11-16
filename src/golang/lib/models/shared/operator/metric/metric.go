package metric

import (
	"github.com/aqueducthq/aqueduct/lib/collections/operator/function"
)

type Metric struct {
	Function function.Function `json:"function"`
}
