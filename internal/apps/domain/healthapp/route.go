package healthapp

import "github.com/go-chi/chi/v5"

func (h *HealthApp) Routes(r chi.Router) {
	r.Get("/liveness", h.liveness)
	r.Get("/readiness", h.readiness)
}
