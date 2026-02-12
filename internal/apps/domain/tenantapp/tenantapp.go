package tenantapp

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type TenantApp struct {
	log  *slog.Logger
	repo ProjectRepository
}

func New(log *slog.Logger, repo ProjectRepository) *TenantApp {
	return &TenantApp{
		log:  log,
		repo: repo,
	}
}

func (h *TenantApp) List(w http.ResponseWriter, r *http.Request) {
	tenantId := chi.URLParam(r, "tenantId")

	projects, err := h.repo.ListProjects(r.Context(), tenantId)
	if err != nil {
		h.log.Error(err.Error())
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, projects)

}

func (h *TenantApp) Create(w http.ResponseWriter, r *http.Request) {

	tenantId := chi.URLParam(r, "tenantId")

	const maxBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := h.repo.CreateProject(r.Context(), tenantId, req.Name)
	if err != nil {
		h.log.Error(err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, project)

}
func (h *TenantApp) Get(w http.ResponseWriter, r *http.Request) {
	tenantId := chi.URLParam(r, "tenantId")
	projectId := chi.URLParam(r, "projectId")

	project, err := h.repo.GetProject(r.Context(), tenantId, projectId)
	if err != nil {
		h.log.Error(err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, project)

}
func (h *TenantApp) Delete(w http.ResponseWriter, r *http.Request) {
	tenantId := chi.URLParam(r, "tenantId")
	projectId := chi.URLParam(r, "projectId")

	err := h.repo.DeleteProject(r.Context(), tenantId, projectId)

	if err != nil {
		h.log.Error(err.Error())
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusNoContent, nil)

}
