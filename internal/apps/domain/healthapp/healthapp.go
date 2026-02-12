package healthapp

import (
	"log/slog"
	"net/http"
)

type HealthApp struct {
	log *slog.Logger
}

func NewHealthApp(log *slog.Logger) *HealthApp {
	return &HealthApp{log: log}
}

func (h *HealthApp) liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *HealthApp) readiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}
