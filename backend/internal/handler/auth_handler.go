package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtService  *service.JWTService
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(
	authService *service.AuthService,
	jwtService *service.JWTService,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtService:  jwtService,
	}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, err := h.authService.Signup(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail),
			errors.Is(err, service.ErrPasswordTooShort):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": userID})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
