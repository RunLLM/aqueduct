package stmt_preparers

import (
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
)

func (s *StandardPreparer) PrepareUpdateExecStateStmt(
	columnAccessPath string,
	status shared.ExecutionStatus,
	timestamp time.Time,
	offset int,
) (fragment string, args []interface{}, err error) {
	timestampField, err := shared.ExecutionTimestampsJsonFieldByStatus(status)
	if err != nil {
		return "", nil, err
	}

	// (likawind) If passing the timestamp object directly to query, it somehow bypass the
	// drivers serialization and results in values that couldn't be deserialized back.
	// The reason is not clear, but using `MarshalText()` to pre-serialize the value
	// appears to solve the issue.
	timestampValue, err := timestamp.MarshalText()
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf(`%s = CAST(
		json_set(
			json_set(%s, '$.status', $%d),
			'$.timestamps.%s',
			$%d
		) AS BLOB)`,
			columnAccessPath,
			columnAccessPath,
			offset+1,
			timestampField,
			offset+2,
		), []interface{}{
			status, string(timestampValue),
		}, nil
}
