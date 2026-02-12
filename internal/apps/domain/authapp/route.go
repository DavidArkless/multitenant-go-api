package authapp

import "github.com/go-chi/chi/v5"

func (a *AuthApp) Routes(r chi.Router) {
	r.Post("/login", a.Login)
}
