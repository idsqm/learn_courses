package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andruho/courses/internal/config"
	"github.com/andruho/courses/internal/handler"
	"github.com/andruho/courses/internal/repository"
	"github.com/andruho/courses/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logLevel := slog.LevelInfo
	if cfg.Debug {
		logLevel = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	handler.SetLogger(log, cfg.Debug)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := connectWithRetry(ctx, cfg.DBURL, 10, log)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := runMigrations(ctx, pool, log); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Repositories
	courses := repository.NewCourseRepository(pool)
	categories := repository.NewCategoryRepository(pool)
	enrollments := repository.NewEnrollmentRepository(pool)
	favorites := repository.NewFavoriteRepository(pool)
	reviews := repository.NewReviewRepository(pool)
	certificates := repository.NewCertificateRepository(pool)
	authors := repository.NewAuthorRepository(pool)
	progress := repository.NewProgressRepository(pool)
	studioRepo := repository.NewStudioRepository(pool)
	lessonContentRepo := repository.NewLessonContentRepository(pool)

	// Services
	courseSvc := service.NewCourseService(courses)
	categorySvc := service.NewCategoryService(categories)
	enrollmentSvc := service.NewEnrollmentService(enrollments, courses)
	favoriteSvc := service.NewFavoriteService(favorites, courses)
	reviewSvc := service.NewReviewService(reviews, courses)
	certificateSvc := service.NewCertificateService(certificates)
	authorSvc := service.NewAuthorService(authors, cfg.AuthServiceURL)
	progressSvc := service.NewProgressService(progress, certificates, enrollments)
	studioSvc := service.NewStudioService(studioRepo)
	contentSvc := service.NewLessonContentService(lessonContentRepo)

	// Router
	router := handler.NewRouter(
		courseSvc, categorySvc, enrollmentSvc, favoriteSvc,
		reviewSvc, certificateSvc, authorSvc, progressSvc,
		studioSvc, contentSvc, cfg.JWTSecret,
	)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("server starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped")
}

func connectWithRetry(ctx context.Context, dbURL string, maxRetries int, log *slog.Logger) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	for i := 0; i < maxRetries; i++ {
		pool, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool, nil
			}
			pool.Close()
		}
		log.Warn("failed to connect to postgres, retrying", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

func gooseUp(raw string) string {
	if i := strings.Index(raw, "-- +goose Down"); i != -1 {
		return raw[:i]
	}
	return raw
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	for _, path := range []string{
		"migrations/20260620001_init.sql",
		"migrations/20260620003_studio.sql",
		"migrations/20260628001_lesson_content.sql",
	} {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, gooseUp(string(raw))); err != nil {
			return err
		}
	}

	seed, err := os.ReadFile("migrations/20260620002_seed.sql")
	if err != nil {
		log.Info("no seed file found, skipping")
		return nil
	}

	var count int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM categories").Scan(&count)
	if count == 0 {
		if _, err := pool.Exec(ctx, string(seed)); err != nil {
			return err
		}
		log.Info("seed data applied")
	}

	return nil
}
