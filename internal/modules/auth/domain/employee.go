package domain

import "time"

type Employee struct {
	ID        int       `db:"id"         json:"id"`
	BranchID  int       `db:"branch_id"  json:"branch_id"`
	NIK       string    `db:"nik"        json:"nik"`
	FullName  string    `db:"full_name"  json:"full_name"`
	Role      string    `db:"role"       json:"role"`
	PinHash   string    `db:"pin_hash"   json:"-"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

