package notification

import "strings"

const (
	tableName = "notification"

	// Notification table column names
	IdColumn          = "id"
	ReceiverIdColumn  = "receiver_id"
	ContentColumn     = "content"
	StatusColumn      = "status"
	LevelColumn       = "level"
	AssociationColumn = "association"
	CreatedAtColumn   = "created_at"
)

// Returns a joined string of all Notification columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			ReceiverIdColumn,
			ContentColumn,
			StatusColumn,
			LevelColumn,
			AssociationColumn,
			CreatedAtColumn,
		},
		",",
	)
}
