package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourusername/broker-backend/internal/domain"
)

// AuthHandler, kimlik doğrulama endpoint'lerini yönetir.
// Public route'lar: JWT middleware gerektirmez.
type AuthHandler struct {
	usecase domain.AuthUsecase
}

// NewAuthHandler, yeni bir AuthHandler örneği döner.
func NewAuthHandler(usecase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: usecase}
}

// ── Request / Response DTOs ───────────────────────────────────────────────────

type registerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string  `json:"token"`
	User  userDTO `json:"user"`
}

// userDTO, hassas alanlar (PasswordHash) çıkarılmış kullanıcı yanıtıdır.
type userDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// Register godoc
//
//	POST /api/auth/register
//	Body: { "email", "password", "first_name", "last_name" }
//	Yanıt: 201 Created → { "token", "user" }
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "geçersiz istek gövdesi"})
		return
	}
	defer r.Body.Close()

	// Alan temizleme
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	// Zorunlu alan kontrolü
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "email, password, first_name ve last_name zorunludur",
		})
		return
	}

	// Minimum şifre uzunluğu
	if len(req.Password) < 8 {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "şifre en az 8 karakter olmalıdır",
		})
		return
	}

	user, token, err := h.usecase.Register(r.Context(), domain.RegisterParams{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  toUserDTO(user),
	})
}

// Login godoc
//
//	POST /api/auth/login
//	Body: { "email", "password" }
//	Yanıt: 200 OK → { "token", "user" }
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "geçersiz istek gövdesi"})
		return
	}
	defer r.Body.Close()

	if req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "email ve password zorunludur",
		})
		return
	}

	user, token, err := h.usecase.Login(r.Context(), strings.TrimSpace(strings.ToLower(req.Email)), req.Password)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  toUserDTO(user),
	})
}

// toUserDTO, domain.User'ı hassas alanlar olmadan DTO'ya dönüştürür.
func toUserDTO(u *domain.User) userDTO {
	return userDTO{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}
