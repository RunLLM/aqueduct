package operator_result

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/logging"
	"github.com/aqueducthq/enterprise/src/golang/lib/collections/utils"
)

type NullExecutionLogs struct {
	logging.ExecutionLogs
	IsNull bool
}

func (n *NullExecutionLogs) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return utils.ValueJsonB(n.ExecutionLogs)
}

func (n *NullExecutionLogs) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	logs := &logging.ExecutionLogs{}
	if err := utils.ScanJsonB(value, logs); err != nil {
		return err
	}

	n.ExecutionLogs, n.IsNull = *logs, false
	return nil
}
