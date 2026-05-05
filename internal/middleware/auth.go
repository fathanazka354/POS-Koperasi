package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/fathanazka354/pos-koperasi/internal/model"
	"github.com/fathanazka354/pos-koperasi/pkg/response"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const EmployeeKey contextKey = "employee"

func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Authorization header required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Unauthorized(w, "Invalid authorization format")
				return
			}

			tokenStr := parts[1]
			claims := &model.EmployeeClaims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				response.Unauthorized(w, "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), EmployeeKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(r *http.Request) *model.EmployeeClaims {
	claims, _ := r.Context().Value(EmployeeKey).(*model.EmployeeClaims)
	return claims
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r)
			if claims == nil {
				response.Unauthorized(w, "Unauthorized")
				return
			}
			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			response.Unauthorized(w, "Insufficient permissions")
		})
	}
}
