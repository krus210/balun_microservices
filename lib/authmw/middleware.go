package authmw

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// HTTPMiddleware создает HTTP middleware для JWT валидации
func HTTPMiddleware(validator *Validator, skipPaths ...string) func(http.Handler) http.Handler {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем пути, которые не требуют аутентификации
			if skipMap[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Извлекаем Authorization header
			authHeader := r.Header.Get(AuthorizationHeader)
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			// Проверяем формат
			if !strings.HasPrefix(authHeader, BearerPrefix) {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

			// Валидируем токен
			claims, err := validator.Validate(r.Context(), tokenString)
			if err != nil {
				if errors.Is(err, ErrTokenExpired) {
					http.Error(w, "token expired", http.StatusUnauthorized)
					return
				}
				if errors.Is(err, ErrInvalidAudience) {
					http.Error(w, "invalid audience", http.StatusForbidden)
					return
				}
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Добавляем user_id в context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)

			// Передаем дальше с обогащенным context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
