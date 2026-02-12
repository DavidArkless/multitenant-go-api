package tenantapp

import (
	"github.com/go-chi/chi/v5"
)

func (h *TenantApp) Routes(r chi.Router) {
	r.Route("/{tenantId}", func(r chi.Router) {
		r.Use(requireTenantAccess)
		r.Get("/projects", h.List)
		r.Get("/projects/{projectId}", h.Get)

		r.Group(func(r chi.Router) {
			r.Use(requireAdminRole)
			r.Post("/projects", h.Create)
			r.Delete("/projects/{projectId}", h.Delete)
		})

	})
}
