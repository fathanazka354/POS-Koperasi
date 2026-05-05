package auth

import "time"

type Member struct {
	ID           int       `db:"id"            json:"id"`
	MemberCode   string    `db:"member_code"   json:"member_code"`
	FullName     string    `db:"full_name"     json:"full_name"`
	Phone        string    `db:"phone"         json:"phone"`
	PointBalance int       `db:"point_balance" json:"point_balance"`
	JoinedAt     time.Time `db:"joined_at"     json:"joined_at"`
	IsActive     bool      `db:"is_active"     json:"is_active"`
}

