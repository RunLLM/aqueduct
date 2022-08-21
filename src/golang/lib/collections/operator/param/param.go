package param

import "github.com/aqueducthq/aqueduct/lib/collections/artifact"

// The value of a parameter must be JSON serializable.
type Param struct {
	Val string `json:"val"`
	Val_Type artifact.Type `json:"val_type"`
}
