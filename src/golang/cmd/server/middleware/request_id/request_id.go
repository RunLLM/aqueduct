package request_id

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aqueducthq/aqueduct/cmd/server/utils"
	"github.com/google/uuid"
)

func WithRequestId() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// We generate an ID if there's any unfortunate issue that
			// uuid generation failed.
			requestIdStr := fmt.Sprintf("error-uuid-generation-%d", time.Now().Unix())
			requestId, err := uuid.NewUUID()
			if err == nil {
				requestIdStr = requestId.String()
			}

			contextWithReqId := context.WithValue(r.Context(), utils.UserRequestIdKey, requestIdStr)
			h.ServeHTTP(w, r.WithContext(contextWithReqId))
		})
	}
}
