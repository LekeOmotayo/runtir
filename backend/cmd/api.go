package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lekeOmotayo/runtir/internal/config"
	"github.com/lekeOmotayo/runtir/internal/db"
	"github.com/lekeOmotayo/runtir/internal/middleware"
	"github.com/lekeOmotayo/runtir/internal/modules/auth"
	"github.com/lekeOmotayo/runtir/internal/modules/document"
	"github.com/lekeOmotayo/runtir/internal/storage"
	"github.com/lekeOmotayo/runtir/internal/utils"
)

type Server struct {
	cfg        *config.Config
	router     *gin.Engine
	pool       *pgxpool.Pool
	queries    *db.Queries
	storage    storage.StorageService
	httpServer *http.Server
}

// NewServer initializes configurations, database pools, storage, middlewares, and routes
func NewServer(cfg *config.Config) (*Server, error) {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 1. Initialize PostgreSQL connection pool using pgx/v5
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	// Ping database to verify connection
	if err := pool.Ping(ctx); err != nil {
		log.Printf("Warning: Database ping failed (%v). Continuing startup...", err)
	} else {
		log.Println("Database connection pool established successfully")
	}

	queries := db.New(pool)

	// 2. Initialize file storage abstraction
	storageService, err := storage.NewLocalStorageService(cfg.UploadDir)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to initialize storage service: %w", err)
	}

	// 3. Initialize Gin engine and core middlewares
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORSMiddleware())

	// Set max multipart memory
	router.MaxMultipartMemory = cfg.MaxUploadSizeMB << 20

	// 4. Register health check
	router.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, "AfriOps API is healthy", gin.H{
			"app_env":   cfg.AppEnv,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	// 5. Versioned API routes group
	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/health", func(c *gin.Context) {
			utils.SuccessResponse(c, http.StatusOK, "AfriOps API v1 is healthy", gin.H{
				"status": "up",
			})
		})

		// Auth Middleware for protected endpoints
		authMW := middleware.AuthMiddleware(cfg, queries)

		// Authentication module
		authService := auth.NewService(cfg, queries)
		authHandler := auth.NewHandler(cfg, authService)
		authHandler.RegisterRoutes(apiV1, authMW)

		// Document ingestion module
		docService := document.NewService(cfg, queries, storageService)
		docHandler := document.NewHandler(docService)
		docHandler.RegisterRoutes(apiV1, authMW)
	}

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		cfg:        cfg,
		router:     router,
		pool:       pool,
		queries:    queries,
		storage:    storageService,
		httpServer: httpServer,
	}, nil
}

// Start begins listening for incoming HTTP traffic and handles graceful shutdown
func (s *Server) Start() error {
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("AfriOps server listening on port %s (%s mode)", s.cfg.Port, s.cfg.AppEnv)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for OS interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Printf("Received signal %v. Initiating graceful shutdown...", sig)
	}

	// Graceful shutdown with 15-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}

	s.Close()
	log.Println("AfriOps server stopped gracefully")
	return nil
}

func (s *Server) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}