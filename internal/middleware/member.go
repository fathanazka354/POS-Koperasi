package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourname/pos-koperasi/pkg/response"
)

const MemberKey contextKey = "member"

type MemberClaims struct {
	MemberID int `json:"member_id"`
	jwt.RegisteredClaims
}

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
			claims := &MemberClaims{}

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

			if claims.Issuer != JWTIssuerMember || claims.MemberID == 0 {
				response.Unauthorized(w, "Invalid member token")
				return
			}

			ctx := context.WithValue(r.Context(), MemberKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetMemberClaims(r *http.Request) *MemberClaims {
	claims, _ := r.Context().Value(MemberKey).(*MemberClaims)
	return claims
}
