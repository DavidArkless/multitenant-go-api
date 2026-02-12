package tenantapp

import (
	"multitenant-go-api/internal/auth"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func requireTenantAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			respondError(w, http.StatusUnauthorized, "user not found")
			return
		}

		tenantID := chi.URLParam(r, "tenantId")

		if user.TenantID != tenantID {
			respondError(w, http.StatusForbidden, "user does not have access to this tenant")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func requireAdminRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			respondError(w, http.StatusUnauthorized, "user not found")
			return

		}

		if user.Role != "admin" {
			respondError(w, http.StatusForbidden, "user does not have admin role")
			return
		}

		next.ServeHTTP(w, r)
	})
}
