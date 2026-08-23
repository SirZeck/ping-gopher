package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SirZeck/ping-gopher/internal/auth"
	"github.com/SirZeck/ping-gopher/internal/config"
	"github.com/SirZeck/ping-gopher/internal/db"
	"gorm.io/gorm"
)

type APIHandler struct {
	DB     *gorm.DB
	Config *config.Config
}

func NewAPIHandler(database *gorm.DB, cfg *config.Config) *APIHandler {
	return &APIHandler{DB: database, Config: cfg}
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string  `json:"token"`
	User  db.User `json:"user"`
}

// SignupHandler handles new tenant user registration.
func (h *APIHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Email == "" || req.Password == "" {
		JSONError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Check if user already exists
	var existing db.User
	if err := h.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		JSONError(w, http.StatusConflict, "User with this email already exists")
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	user := db.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to create user account")
		return
	}

	token, err := auth.GenerateJWTToken(user.ID, user.Email, h.Config.JWTSecret, 24*time.Hour)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	JSONResponse(w, http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// LoginHandler authenticates user credentials and returns a JWT token.
func (h *APIHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var user db.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		JSONError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		JSONError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := auth.GenerateJWTToken(user.ID, user.Email, h.Config.JWTSecret, 24*time.Hour)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	JSONResponse(w, http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}
