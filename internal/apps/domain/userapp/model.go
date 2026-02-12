package userapp

import "time"

type User struct {
	Id        string    `db:"id"`
	TenantID  string    `db:"tenant_id"`
	Email     string    `db:"email"`
	Password  string    `db:"password_hash"`
	Role      string    `db:"role"` // admin or viewer
	CreatedAt time.Time `db:"created_at"`
}
