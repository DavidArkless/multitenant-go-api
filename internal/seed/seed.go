package seed

import (
	"context"
	"multitenant-go-api/internal/auth"
	"time"

	"github.com/jmoiron/sqlx"
)

type Options struct {
	Force bool
}

func SeedIfEmpty(ctx context.Context, db *sqlx.DB, opt Options) error {
	empty, err := isEmpty(ctx, db)
	if err != nil {
		return err
	}

	if !empty && !opt.Force {
		return nil
	}

	return seedAll(ctx, db)
}

func isEmpty(ctx context.Context, db *sqlx.DB) (bool, error) {
	var count int
	err := db.GetContext(ctx, &count, `SELECT COUNT(1) FROM tenants`)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func seedAll(ctx context.Context, db *sqlx.DB) error {

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO tenants (id, name, created_at) VALUES
		('tenant-1', 'Tenant One', ?),
		('tenant-2', 'Tenant Two', ?)
	`, now, now)
	if err != nil {
		return err
	}

	password, err := auth.HashPassword("secret")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, role, created_at) VALUES
		('user-1', 'tenant-1', 'admin@tenant1.com',  ?, 'admin',  ?), 
		('user-2', 'tenant-1', 'viewer@tenant1.com', ?, 'viewer', ?),
		('user-3', 'tenant-2', 'admin@tenant2.com',  ?, 'admin',  ?)
	`, password, now, password, now, password, now)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO projects (id, tenant_id, name, created_at) VALUES
		('p1', 'tenant-1', 'Tenant 1 - Project A', ?),
		('p2', 'tenant-1', 'Tenant 1 - Project B', ?),
		('p3', 'tenant-2', 'Tenant 2 - Project A', ?)
	`, now, now, now)
	if err != nil {
		return err
	}

	return tx.Commit()

}
