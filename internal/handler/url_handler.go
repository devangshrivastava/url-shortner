package handler

import (
	"net/http"

	"url-shortener/internal/service"
	"github.com/gin-gonic/gin"
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
		URL string `json:"url"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	code, err := h.service.Shorten(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "url is required",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      code,
		"short_url": "http://localhost:8080/" + code,
	})
}

func (h *URLHandler) RedirectURL(c *gin.Context) {
	code := c.Param("code")

	longURL, err := h.service.Resolve(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "short url not found",
		})
		return
	}

	c.Redirect(http.StatusFound, longURL)
}