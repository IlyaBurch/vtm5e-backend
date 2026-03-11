package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/vtm5e/backend/config"
	"github.com/vtm5e/backend/internal/auth"
	"github.com/vtm5e/backend/internal/character"
	"github.com/vtm5e/backend/internal/db"
	"github.com/vtm5e/backend/internal/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	migrationsDir := migrationsPath()
	if err := db.RunMigrations(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	log.Println("migrations applied")

	userRepo := user.NewRepository(pool)
	userService := user.NewService(userRepo)
	authHandler := auth.NewHandler(userService, cfg.JWTSecret)

	charRepo := character.NewRepository(pool)
	charService := character.NewService(charRepo)
	charHandler := character.NewHandler(charService)

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestLogger(&chimiddleware.DefaultLogFormatter{
		Logger:  log.New(os.Stdout, "", log.LstdFlags),
		NoColor: true,
	}))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.AllowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	r.Use(corsMiddleware.Handler)

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		r.Route("/characters", func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))
			r.Get("/", charHandler.List)
			r.Post("/", charHandler.Create)
			r.Get("/{id}", charHandler.Get)
			r.Put("/{id}", charHandler.Update)
			r.Delete("/{id}", charHandler.Delete)
		})
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("server starting on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func migrationsPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "migrations"
	}
	// cmd/server/main.go -> ../../migrations
	return filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
}
