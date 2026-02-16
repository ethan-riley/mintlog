package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/felipemonteiro/mintlog/pkg/apierror"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"request_id", GetRequestID(r.Context()),
				)
				apierror.Write(w, apierror.Internal("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
