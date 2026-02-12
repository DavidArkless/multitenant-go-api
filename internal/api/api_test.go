package api

import (
	"bytes"
	"context"
	"encoding/json"
	"multitenant-go-api/internal/apps/domain/authapp"
	"multitenant-go-api/internal/apps/domain/healthapp"
	"multitenant-go-api/internal/apps/domain/tenantapp"
	"multitenant-go-api/internal/apps/domain/userapp"
	"multitenant-go-api/internal/auth"
	"multitenant-go-api/internal/config"
	"multitenant-go-api/internal/seed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func setupTestAPI(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to in-memory db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE tenants (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('admin', 'viewer')),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_users_tenant ON users(tenant_id);
		CREATE TABLE projects (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_projects_tenant ON projects(tenant_id);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	if err := seed.SeedIfEmpty(context.Background(), db, seed.Options{Force: true}); err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	cfg := config.Api{
		ApiAddr: ":0",
	}

	tenantRepo := tenantapp.NewSQLiteProjectRepo(db)
	userRepo := userapp.NewSQLiteUserRepo(db)
	jwtService := auth.NewJWTService("test-secret")

	apps := Apps{
		HealthApp: healthapp.NewHealthApp(nil),
		TenantApp: tenantapp.New(nil, tenantRepo),
		AuthApp:   authapp.New(nil, jwtService, userRepo),
	}

	authentication := Authentication{
		JwtService: jwtService,
	}

	apiServer := NewApi(cfg, nil, apps, authentication)
	server := httptest.NewServer(apiServer.Handler())

	cleanup := func() {
		server.Close()
		db.Close()
	}

	return server, cleanup
}

func getToken(t *testing.T, serverURL, email, password string) string {
	t.Helper()

	body := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed with status: %d", resp.StatusCode)
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	return loginResp.Token
}

func TestCrossTenantIsolation(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()

	// Login as tenant-1 admin
	token := getToken(t, server.URL, "admin@tenant1.com", "secret")

	// Try to access tenant-2's projects (should be forbidden)
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/tenants/tenant-2/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403 Forbidden, got %d", resp.StatusCode)
	}
}

func TestViewerCannotDelete(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()

	// Login as tenant-1 viewer
	token := getToken(t, server.URL, "viewer@tenant1.com", "secret")

	// Try to delete a project (should be forbidden)
	req, _ := http.NewRequest("DELETE", server.URL+"/api/v1/tenants/tenant-1/projects/p1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403 Forbidden, got %d", resp.StatusCode)
	}
}

func TestViewerCanRead(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()

	// Login as tenant-1 viewer
	token := getToken(t, server.URL, "viewer@tenant1.com", "secret")

	// List projects (should succeed)
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/tenants/tenant-1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}

	var projects []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(projects) == 0 {
		t.Error("expected some projects, got none")
	}
}

func TestAdminCanCreateAndDelete(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()

	// Login as tenant-1 admin
	token := getToken(t, server.URL, "admin@tenant1.com", "secret")

	// Create a project
	createBody := map[string]string{"name": "Test Project"}
	jsonBody, _ := json.Marshal(createBody)

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/tenants/tenant-1/projects", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	var project map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	projectID := project["id"].(string)

	// Delete the project
	req, _ = http.NewRequest("DELETE", server.URL+"/api/v1/tenants/tenant-1/projects/"+projectID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204 No Content, got %d", resp.StatusCode)
	}
}

func TestUnauthenticatedRequestFails(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()

	// Try to access projects without token
	resp, err := http.Get(server.URL + "/api/v1/tenants/tenant-1/projects")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestInvalidTokenFails(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/tenants/tenant-1/projects", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", resp.StatusCode)
	}
}
