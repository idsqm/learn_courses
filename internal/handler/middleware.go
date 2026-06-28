package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/andruho/courses/internal/domain"
)

type ctxKey string

const userIDKey ctxKey = "user_id"
const authorIDKey ctxKey = "author_id"

func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

func AuthorIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(authorIDKey).(string)
	return v
}

type AuthorResolver func(ctx context.Context, userID string) (string, error)

func AuthorMiddleware(resolve AuthorResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserIDFromContext(r.Context())
			if userID == "" {
				writeError(w, domain.ErrAccessTokenInvalid)
				return
			}
			authorID, err := resolve(r.Context(), userID)
			if err != nil {
				writeError(w, err)
				return
			}
			ctx := context.WithValue(r.Context(), authorIDKey, authorID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeError(w, domain.ErrAccessTokenInvalid)
				return
			}

			tokenStr, ok := strings.CutPrefix(header, "Bearer ")
			if !ok {
				writeError(w, domain.ErrAccessTokenInvalid)
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil {
				if strings.Contains(err.Error(), "expired") {
					writeError(w, domain.ErrAccessTokenExpired)
				} else {
					writeError(w, domain.ErrAccessTokenInvalid)
				}
				return
			}

			sub, err := token.Claims.GetSubject()
			if err != nil || sub == "" {
				writeError(w, domain.ErrAccessTokenInvalid)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenStr, ok := strings.CutPrefix(header, "Bearer ")
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			sub, _ := token.Claims.GetSubject()
			if sub != "" {
				ctx := context.WithValue(r.Context(), userIDKey, sub)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
