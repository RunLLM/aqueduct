package _000011_exec_state_column_backfill

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

type Metadata struct {
	Error string            `json:"error"`
	Logs  map[string]string `json:"logs"`
}

type NullMetadata struct {
	Metadata
	IsNull bool
}

func (m *Metadata) Value() (driver.Value, error) {
	return utils.ValueJSONB(*m)
}

func (m *Metadata) Scan(value interface{}) error {
	return utils.ScanJSONB(value, m)
}

func (n *NullMetadata) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.Metadata).Value()
}

func (n *NullMetadata) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	metadata := &Metadata{}
	if err := metadata.Scan(value); err != nil {
		return err
	}

	n.Metadata, n.IsNull = *metadata, false
	return nil
}

type Logs struct {
	Stdout string `json:"stdout"`
	StdErr string `json:"stderr"`
}

type Error struct {
	Context string `json:"context"`
	Tip     string `json:"tip"`
}

type ExecutionState struct {
	UserLogs    *Logs                  `json:"user_logs"`
	Status      shared.ExecutionStatus `json:"status"`
	FailureType *shared.FailureType    `json:"failure_type"`
	Error       *Error                 `json:"error"`
}

func (e *ExecutionState) Value() (driver.Value, error) {
	return utils.ValueJSONB(*e)
}

func (e *ExecutionState) Scan(value interface{}) error {
	return utils.ScanJSONB(value, e)
}

type NullExecutionState struct {
	ExecutionState
	IsNull bool
}

func (n *NullExecutionState) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.ExecutionState).Value()
}

func (n *NullExecutionState) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	logs := &ExecutionState{}
	if err := logs.Scan(value); err != nil {
		return err
	}

	n.ExecutionState, n.IsNull = *logs, false
	return nil
}
