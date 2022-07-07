package header_stripping

import (
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
)

var usefulHeaders = map[string]bool{
	routes.ContentTypeHeader:        true,
	routes.ApiKeyHeader:             true,
	routes.SdkClientVersionHeader:   true,
	routes.IntegrationNameHeader:    true,
	routes.IntegrationServiceHeader: true,
	routes.IntegrationConfigHeader:  true,
	routes.TableNameHeader:          true,
}

func StripHeader() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			toRemove := []string{}
			// Loop over header names
			for name := range r.Header {
				if _, ok := usefulHeaders[name]; !ok {
					toRemove = append(toRemove, name)
				}
			}

			for _, header := range toRemove {
				r.Header.Del(header)
			}

			h.ServeHTTP(w, r)
		})
	}
}
