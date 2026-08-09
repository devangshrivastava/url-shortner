package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"
)

func main() {
	// Load variables from .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Read PostgreSQL connection URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Basic context for startup database operations
	ctx := context.Background()

	// Create PostgreSQL connection pool
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal("failed to create database pool:", err)
	}
	defer db.Close()

	// Check that PostgreSQL is actually reachable
	if err := db.Ping(ctx); err != nil {
		log.Fatal("failed to connect to PostgreSQL:", err)
	}

	log.Println("Connected to PostgreSQL")

	// Dependency setup
	urlRepo := repository.NewPostgresURLRepository(db)
	urlService := service.NewURLService(urlRepo)
	urlHandler := handler.NewURLHandler(urlService)

	// Gin router
	r := gin.Default()

	r.POST("/shorten", urlHandler.ShortenURL)
	r.GET("/:code", urlHandler.RedirectURL)
	r.GET("/analytics/:code", urlHandler.GetAnalytics)

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}