package api

import (
	"log/slog"
	"multitenant-go-api/internal/apps/domain/authapp"
	"multitenant-go-api/internal/apps/domain/healthapp"
	"multitenant-go-api/internal/apps/domain/tenantapp"
	"multitenant-go-api/internal/auth"
	"multitenant-go-api/internal/config"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Apps struct {
	HealthApp *healthapp.HealthApp
	TenantApp *tenantapp.TenantApp // future
	AuthApp   *authapp.AuthApp
}

type Authentication struct {
	JwtService *auth.JWTService
}

type Api struct {
	server *http.Server
	log    *slog.Logger
	config config.Api
	apps   Apps
	auth   Authentication
}

func NewApi(cfg config.Api, log *slog.Logger, apps Apps, authentication Authentication) *Api {
	api := &Api{
		config: cfg,
		log:    log,
		apps:   apps,
		auth:   authentication,
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	// In a real project would create a middleware logger
	router.Use(middleware.Logger)

	router.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Route("/health", api.apps.HealthApp.Routes)
			r.Route("/auth", api.apps.AuthApp.Routes)
		})

		r.Group(func(r chi.Router) {
			r.Use(auth.JWTMiddleware(api.auth.JwtService))
			r.Route("/tenants", api.apps.TenantApp.Routes)

		})

	})

	api.server = &http.Server{
		Addr:         cfg.ApiAddr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return api
}

func (a *Api) Start() error {
	a.log.Info("starting API server", "addr", a.config.ApiAddr)
	return a.server.ListenAndServe()
}

func (a *Api) Handler() http.Handler {
	return a.server.Handler
}
