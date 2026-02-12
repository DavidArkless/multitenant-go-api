package tenantapp

import (
	"encoding/json"
	"net/http"
	"time"
)

type Project struct {
	Id        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	TenantID  string    `db:"tenant_id" json:"tenant_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CreateProjectRequest struct {
	Name string `json:"name"`
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
