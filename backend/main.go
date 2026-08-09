package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

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

	// Create the target database if it does not exist yet. PostgreSQL cannot
	// create a database while connected to that same database, so this first
	// connects to the built-in "postgres" database.
	if err := ensureDatabase(ctx, databaseURL); err != nil {
		log.Fatal("failed to ensure PostgreSQL database exists:", err)
	}

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

	if err := repository.InitializePostgresSchema(ctx, db); err != nil {
		log.Fatal("failed to initialize PostgreSQL schema:", err)
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

func ensureDatabase(ctx context.Context, databaseURL string) error {
	dbURL, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	databaseName := strings.TrimPrefix(dbURL.Path, "/")
	if databaseName == "" {
		return fmt.Errorf("DATABASE_URL does not specify a database name")
	}

	adminURL := *dbURL
	adminURL.Path = "/postgres"
	adminURL.RawPath = ""

	adminDB, err := pgxpool.New(ctx, adminURL.String())
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL admin database: %w", err)
	}
	defer adminDB.Close()

	if err := adminDB.Ping(ctx); err != nil {
		return fmt.Errorf("connect to PostgreSQL admin database: %w", err)
	}

	var exists bool
	err = adminDB.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)",
		databaseName,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// Database identifiers cannot be passed as query parameters. Quote the
	// name safely before placing it in CREATE DATABASE.
	quotedName := `"` + strings.ReplaceAll(databaseName, `"`, `""`) + `"`
	_, err = adminDB.Exec(ctx, "CREATE DATABASE "+quotedName)
	if err != nil {
		// Another application instance may have created it between the check
		// above and CREATE DATABASE. Treat that case as success.
		var nowExists bool
		checkErr := adminDB.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)",
			databaseName,
		).Scan(&nowExists)
		if checkErr == nil && nowExists {
			return nil
		}
		return err
	}

	return nil
}
