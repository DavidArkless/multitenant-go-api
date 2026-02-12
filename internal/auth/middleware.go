package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const userKey ctxKey = "user"

type UserContext struct {
	UserID   string
	TenantID string
	Role     string
}

func JWTMiddleware(jwt *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ParseToken(parts[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, &UserContext{
				UserID:   claims.UserID,
				TenantID: claims.TenantID,
				Role:     claims.Role,
			})

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func UserFromContext(ctx context.Context) (*UserContext, bool) {
	u, ok := ctx.Value(userKey).(*UserContext)
	return u, ok
}

func WithUser(ctx context.Context, u *UserContext) context.Context {
	return context.WithValue(ctx, userKey, u)
}
