package auth

import (
	"net/http"

	"github.com/felipemonteiro/mintlog/internal/tenant"
	"github.com/felipemonteiro/mintlog/pkg/apierror"
)

func Middleware(resolver *KeyResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				apierror.Write(w, apierror.Unauthorized("missing X-API-Key header"))
				return
			}

			info, err := resolver.Resolve(r.Context(), key)
			if err != nil {
				apierror.Write(w, apierror.Unauthorized("invalid or expired API key"))
				return
			}

			ctx := tenant.WithInfo(r.Context(), info)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := tenant.FromContext(r.Context())
			if info == nil {
				apierror.Write(w, apierror.Unauthorized("not authenticated"))
				return
			}
			if !HasScope(info.Scopes, scope) {
				apierror.Write(w, apierror.Forbidden("insufficient scope: "+scope))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
