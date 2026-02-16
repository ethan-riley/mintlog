package middleware

import (
	"net/http"

	rediscache "github.com/felipemonteiro/mintlog/internal/storage/redis"
	"github.com/felipemonteiro/mintlog/internal/tenant"
	"github.com/felipemonteiro/mintlog/pkg/apierror"
)

func RateLimit(limiter *rediscache.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := tenant.FromContext(r.Context())
			if info == nil {
				next.ServeHTTP(w, r)
				return
			}

			allowed, _, err := limiter.Allow(r.Context(), info.ID.String(), int(info.RateLimit))
			if err != nil {
				// If rate limiter fails, allow the request but log
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				apierror.Write(w, apierror.TooManyRequests("rate limit exceeded"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
