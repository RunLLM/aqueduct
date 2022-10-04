package request

import (
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
)

// ParseExportUserFriendlyFromRequest parses whether this export request is only for
// user-friendly code.
func ParseExportUserFriendlyFromRequest(r *http.Request) bool {
	userFriendly := r.Header.Get(routes.ExportFnUserFriendlyHeader)
	return userFriendly != ""
}
