package operator_result

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
)

type NullExecutionLogs struct {
	shared.ExecutionLogs
	IsNull bool
}

func (n *NullExecutionLogs) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.ExecutionLogs).Value()
}

func (n *NullExecutionLogs) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	logs := &shared.ExecutionLogs{}
	if err := logs.Scan(value); err != nil {
		return err
	}

	n.ExecutionLogs, n.IsNull = *logs, false
	return nil
}
