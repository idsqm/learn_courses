package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/andruho/courses/internal/service"
)

func NewRouter(
	courseSvc service.CourseService,
	categorySvc service.CategoryService,
	enrollmentSvc service.EnrollmentService,
	favoriteSvc service.FavoriteService,
	reviewSvc service.ReviewService,
	certificateSvc service.CertificateService,
	authorSvc service.AuthorService,
	progressSvc service.ProgressService,
	jwtSecret string,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	course := NewCourseHandler(courseSvc)
	category := NewCategoryHandler(categorySvc)
	enrollment := NewEnrollmentHandler(enrollmentSvc)
	favorite := NewFavoriteHandler(favoriteSvc)
	review := NewReviewHandler(reviewSvc)
	certificate := NewCertificateHandler(certificateSvc)
	author := NewAuthorHandler(authorSvc)
	progress := NewProgressHandler(progressSvc)
	stats := NewStatsHandler(progressSvc)

	r.Route("/api/v1", func(r chi.Router) {
		// Public
		r.Get("/courses", course.List)
		r.Get("/courses/featured", course.Featured)
		r.Get("/courses/{id}", course.GetByID)
		r.Get("/categories", category.List)
		r.Get("/authors/{id}", author.GetByID)

		// Optional auth
		r.Group(func(r chi.Router) {
			r.Use(OptionalAuthMiddleware(jwtSecret))
			r.Get("/courses/recommended", course.Recommended)
		})

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(jwtSecret))

			r.Post("/courses/{id}/enroll", enrollment.Enroll)
			r.Get("/enrollments", enrollment.List)

			r.Post("/courses/{id}/favorite", favorite.Add)
			r.Delete("/courses/{id}/favorite", favorite.Remove)
			r.Get("/favorites", favorite.List)

			r.Post("/courses/{id}/lessons/{lessonID}/complete", progress.CompleteLesson)
			r.Get("/courses/{id}/progress", progress.GetCourseProgress)

			r.Post("/courses/{id}/reviews", review.Create)

			r.Get("/users/me/stats", stats.GetUserStats)

			r.Get("/certificates", certificate.List)
			r.Get("/certificates/{id}", certificate.GetByID)

			r.Post("/authors/apply", author.Apply)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return r
}
