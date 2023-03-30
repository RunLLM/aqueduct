package v2

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage_migration"
	"github.com/dropbox/godropbox/errors"
)

/*
This file should map directory to src/ui/common/src/handlers/ListStorageMigrations.tsx

Route: /v2/storage-migrations
Method: GET
Request:
	Headers:
		`api-key`:
			User's API Key
		`status`:
			Optional filter that returns only those storage migrations with the given status.
	    	Defaults to all statuses.
		`limit`:
			Optional limit on the number of storage migrations returned. Defaults to all of them.
		`completed-since`:
			Optional unix timestamp. If set,gg we wil only return storage migrations that have
			completed since this time.

	We always return storage migrations in descending chronological order (by start time).
	The order the filters are applied in is: status, completed-since, then limit.
*/

type ListStorageMigrationsHandler struct {
	handler.GetHandler

	Database             database.Database
	StorageMigrationRepo repos.StorageMigration
}

type listStorageMigrationsArgs struct {
	*aq_context.AqContext

	// See the route description above for what each of these filters mean.
	status         *string
	limit          int
	completedSince *time.Time
}

func (*ListStorageMigrationsHandler) Headers() []string {
	return []string{
		routes.StorageMigrationFilterStatusHeader,
		routes.StorageMigrationLimitHeader,
		routes.StorageMigrationCompletedSinceHeader,
	}
}

func (*ListStorageMigrationsHandler) Name() string {
	return "ListStorageMigrations"
}

func (h *ListStorageMigrationsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	limit := -1
	if limitVal := r.Header.Get(routes.StorageMigrationLimitHeader); len(limitVal) > 0 {
		limit, err = strconv.Atoi(limitVal)
		if err != nil {
			return nil, http.StatusBadRequest, errors.Wrap(err, "Invalid limit header.")
		}
	}

	var status *string
	if statusVal := r.Header.Get(routes.StorageMigrationFilterStatusHeader); len(statusVal) > 0 {
		status = &statusVal
	}

	var completedSince *time.Time
	if completedSinceVal := r.Header.Get(routes.StorageMigrationCompletedSinceHeader); len(completedSinceVal) > 0 {
		completedSinceTS, err := strconv.Atoi(completedSinceVal)
		if err != nil {
			return nil, http.StatusBadRequest, errors.Wrap(err, "Invalid completed-since header.")
		}
		completedSinceTime := time.Unix(int64(completedSinceTS), 0)
		completedSince = &completedSinceTime
	}

	return &listStorageMigrationsArgs{
		AqContext:      aqContext,
		limit:          limit,
		status:         status,
		completedSince: completedSince,
	}, http.StatusOK, nil
}

func (h *ListStorageMigrationsHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*listStorageMigrationsArgs)

	migrations, err := storage_migration.ListStorageMigrations(ctx, args.status, args.limit, args.completedSince, h.StorageMigrationRepo, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Failed to list storage migrations.")
	}
	return migrations, http.StatusOK, nil
}
