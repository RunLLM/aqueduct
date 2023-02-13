package metric

import "github.com/aqueducthq/aqueduct/lib/models/shared/operator/function"

type Metric struct {
	Function function.Function `json:"function"`
}
