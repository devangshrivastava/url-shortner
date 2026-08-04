package main

import (
	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	urlRepo := repository.NewMemoryURLRepository()
	urlService := service.NewURLService(urlRepo)
	urlHandler := handler.NewURLHandler(urlService)

	r.POST("/shorten", urlHandler.ShortenURL)
	r.GET("/:code", urlHandler.RedirectURL)

	r.Run(":8080")
}