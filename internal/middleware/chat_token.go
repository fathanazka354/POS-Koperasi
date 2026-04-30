package middleware

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ChatPrincipal identitas untuk koneksi WebSocket chat (karyawan atau member).
type ChatPrincipal struct {
	IsEmployee bool
	EmployeeID int
	BranchID   int
	Role       string
	MemberID   int
}

// ParseChatToken mem-parse Bearer token karyawan atau member (query ?token= pada WS).
func ParseChatToken(tokenStr, secret string) (*ChatPrincipal, error) {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}

	var mc jwt.MapClaims
	_, err := jwt.ParseWithClaims(tokenStr, &mc, keyFunc)
	if err != nil {
		return nil, err
	}

	iss, _ := mc["iss"].(string)
	switch iss {
	case JWTIssuerEmployee:
		eid := intFromClaim(mc, "employee_id")
		if eid == 0 {
			return nil, fmt.Errorf("invalid employee token")
		}
		return &ChatPrincipal{
			IsEmployee: true,
			EmployeeID: eid,
			BranchID:   intFromClaim(mc, "branch_id"),
			Role:       strFromClaim(mc, "role"),
		}, nil
	case JWTIssuerMember:
		mid := intFromClaim(mc, "member_id")
		if mid == 0 {
			return nil, fmt.Errorf("invalid member token")
		}
		return &ChatPrincipal{MemberID: mid}, nil
	default:
		return nil, fmt.Errorf("unknown or missing token issuer")
	}
}

func intFromClaim(mc jwt.MapClaims, key string) int {
	v, ok := mc[key]
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	default:
		return 0
	}
}

func strFromClaim(mc jwt.MapClaims, key string) string {
	v, _ := mc[key].(string)
	return v
}
