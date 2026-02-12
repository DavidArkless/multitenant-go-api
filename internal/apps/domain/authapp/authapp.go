package authapp

import (
	"encoding/json"
	"log/slog"
	"multitenant-go-api/internal/apps/domain/userapp"
	"net/http"

	"multitenant-go-api/internal/auth"
)

type AuthApp struct {
	log      *slog.Logger
	jwt      *auth.JWTService
	userRepo userapp.UserRepository
}

func New(log *slog.Logger, jwt *auth.JWTService, userRepo userapp.UserRepository) *AuthApp {
	return &AuthApp{
		log:      log,
		jwt:      jwt,
		userRepo: userRepo,
	}
}

func (a *AuthApp) Login(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := a.userRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid request body")
		return
	}

	if auth.VerifyPassword(user.Password, req.Password) != nil {
		respondError(w, http.StatusUnauthorized, "invalid request body")
		return
	}

	token, err := a.jwt.GenerateToken(user.Id, user.TenantID, user.Role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := LoginResponse{token}

	respondJSON(w, http.StatusOK, response)

}
