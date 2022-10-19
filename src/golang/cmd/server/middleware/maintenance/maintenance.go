package maintenance

import (
	"net/http"
	"sync/atomic"

	"github.com/aqueducthq/aqueduct/cmd/server/response"
)

// Check validates that the server is not currently under system maintenance.
// If it is, it returns an error response to the client.
func Check(underMaintenance *atomic.Value) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if underMaintenance.Load().(bool) {
				// The server is currently under maintenance
				response.SendErrorResponse(w, "The server is currently unavailable due to system maintenance.", http.StatusServiceUnavailable)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}
