package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"url-shortener/internal/model"
	"url-shortener/internal/service"
)

type URLHandler struct {
	service *service.URLService
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{
		service: service,
	}
}

func (h *URLHandler) ShortenURL(c *gin.Context) {
	var req struct {
		URL          string `json:"url"`
		CustomCode   string `json:"custom_code"`
		ExpiresAt    string `json:"expires_at"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	code, err := h.service.Shorten(req.URL, req.CustomCode, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      code,
		"short_url": "http://localhost:8080/" + code,
	})
}

// func (h *URLHandler) RedirectURL(c *gin.Context) {
// 	code := c.Param("code")

// 	longURL, err := h.service.Resolve(code)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"error": "short url not found",
// 		})
// 		return
// 	}

// 	c.Redirect(http.StatusFound, longURL)
// }

func (h *URLHandler) RedirectURL(c *gin.Context) {
	code := c.Param("code")

	longURL, err := h.service.Resolve(code)
	if err != nil {
		status := http.StatusNotFound

		if err == service.ErrURLExpired {
			status = http.StatusGone
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.service.TrackClick(model.Click{
		Code:      code,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
	})

	c.Redirect(http.StatusFound, longURL)
}

func (h *URLHandler) GetAnalytics(c *gin.Context) {
	code := c.Param("code")

	analytics := h.service.Analytics(code)

	c.JSON(http.StatusOK, analytics)
}