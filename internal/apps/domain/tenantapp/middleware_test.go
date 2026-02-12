package tenantapp

import (
	"context"
	"multitenant-go-api/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRequireTenantAccess(t *testing.T) {
	tests := []struct {
		name       string
		user       *auth.UserContext
		urlTenant  string
		wantStatus int
	}{
		{
			name: "same tenant allowed",
			user: &auth.UserContext{
				UserID:   "u1",
				TenantID: "tenant-1",
				Role:     "admin",
			},
			urlTenant:  "tenant-1",
			wantStatus: http.StatusOK,
		},
		{
			name: "different tenant forbidden",
			user: &auth.UserContext{
				UserID:   "u1",
				TenantID: "tenant-1",
				Role:     "admin",
			},
			urlTenant:  "tenant-2",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "no user unauthorized",
			user:       nil,
			urlTenant:  "tenant-1",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			mw := requireTenantAccess(next)

			req := httptest.NewRequest(http.MethodGet, "/tenants/"+tt.urlTenant+"/projects", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("tenantId", tt.urlTenant)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req = req.WithContext(withUser(req.Context(), tt.user))

			rec := httptest.NewRecorder()

			mw.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})

	}

}

func TestRequireAdminRole(t *testing.T) {
	tests := []struct {
		name       string
		user       *auth.UserContext
		wantStatus int
	}{
		{
			name: "admin allowed",
			user: &auth.UserContext{
				UserID:   "u1",
				TenantID: "tenant-1",
				Role:     "admin",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "viewer forbidden",
			user: &auth.UserContext{
				UserID:   "u2",
				TenantID: "tenant-1",
				Role:     "viewer",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "no user unauthorized",
			user:       nil,
			wantStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			mw := requireAdminRole(next)

			req := httptest.NewRequest(http.MethodGet, "/anything", nil)
			req = req.WithContext(withUser(req.Context(), tt.user))

			rec := httptest.NewRecorder()

			mw.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

		})
	}

}

func withUser(ctx context.Context, user *auth.UserContext) context.Context {
	if user == nil {
		return ctx
	}
	return auth.WithUser(ctx, user)
}
