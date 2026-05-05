package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/fathanazka354/pos-koperasi/internal/model"
	"github.com/fathanazka354/pos-koperasi/pkg/response"
	"github.com/golang-jwt/jwt/v5"
)

const MemberKey contextKey = "member"

func MemberJWTAuth(jwtSecret string) func(http.Handler) http.Handler {
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
			claims := &model.MemberClaims{}

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

			if claims.Issuer != model.JWTIssuerMember || claims.MemberID == 0 {
				response.Unauthorized(w, "Invalid member token")
				return
			}

			ctx := context.WithValue(r.Context(), MemberKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetMemberClaims(r *http.Request) *model.MemberClaims {
	claims, _ := r.Context().Value(MemberKey).(*model.MemberClaims)
	return claims
}

// ParseMemberToken mem-parse token string menjadi MemberClaims (untuk WebSocket query param).
func ParseMemberToken(tokenStr, jwtSecret string) (*model.MemberClaims, error) {
	claims := &model.MemberClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims.Issuer != model.JWTIssuerMember || claims.MemberID == 0 {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
