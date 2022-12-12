package repos

import "github.com/dropbox/godropbox/errors"

var ErrInvalidPendingTimestamp = errors.New("Execution state doesn't have a valid pending_at timestamp.")
