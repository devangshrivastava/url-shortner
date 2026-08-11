package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/model"
	"url-shortener/internal/service"
)

type URLHandler struct {
	service *service.URLService
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{service: service}
}

func (h *URLHandler) ShortenURL(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req struct {
		URL        string `json:"url"`
		CustomCode string `json:"custom_code"`
		ExpiresAt  string `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	code, err := h.service.Shorten(
		c.Request.Context(),
		userID,
		req.URL,
		req.CustomCode,
		req.ExpiresAt,
	)
	if err != nil {
		writeURLError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":      code,
		"short_url": "http://localhost:8080/" + code,
	})
}

func (h *URLHandler) RedirectURL(c *gin.Context) {
	code := c.Param("code")

	longURL, err := h.service.Resolve(c.Request.Context(), code)
	if err != nil {
		writeURLError(c, err)
		return
	}

	if err := h.service.TrackClick(c.Request.Context(), model.Click{
		Code:      code,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
	}); err != nil {
		log.Printf("failed to track click for %q: %v", code, err)
	}

	c.Redirect(http.StatusFound, longURL)
}

func (h *URLHandler) GetAnalytics(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	analytics, err := h.service.GetAnalytics(c.Request.Context(), userID, c.Param("code"))
	if err != nil {
		writeURLError(c, err)
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (h *URLHandler) ListMyURLs(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	urls, err := h.service.ListURLs(c.Request.Context(), userID)
	if err != nil {
		writeURLError(c, err)
		return
	}

	c.JSON(http.StatusOK, urls)
}

func (h *URLHandler) UpdateURL(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req struct {
		LongURL   *string `json:"long_url"`
		ExpiresAt *string `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	url, err := h.service.UpdateURL(
		c.Request.Context(),
		userID,
		c.Param("code"),
		req.LongURL,
		req.ExpiresAt,
	)
	if err != nil {
		writeURLError(c, err)
		return
	}

	c.JSON(http.StatusOK, url)
}

func (h *URLHandler) DeleteURL(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteURL(c.Request.Context(), userID, c.Param("code")); err != nil {
		writeURLError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func authenticatedUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, false
	}

	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, false
	}

	return id, true
}

func writeURLError(c *gin.Context, err error) {
	switch err {
	case service.ErrInvalidURL, service.ErrInvalidExpiry:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case service.ErrURLCodeExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case service.ErrURLNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case service.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case service.ErrURLExpired:
		c.JSON(http.StatusGone, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
