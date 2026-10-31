package middleware

import (
	"context"
	"net/http"
	"strings"

	sharedjwt "shared/jwt"
)

// RequireRole rejects requests whose JWT claims don't match the given role.
// Must be used after JWTMiddleware so that claims are already in context.
func RequireRole(role int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := sharedjwt.GetClaimsFromCtx(r.Context())
			if claims == nil || claims.Role != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func JWTMiddleware(cfg *sharedjwt.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := sharedjwt.ParseAndValidateAccessToken(cfg, tokenStr)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), sharedjwt.CtxKeyClaims{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
