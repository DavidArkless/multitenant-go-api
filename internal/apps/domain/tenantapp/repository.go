package tenantapp

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ProjectRepository interface {
	ListProjects(ctx context.Context, tenantID string) ([]Project, error)
	CreateProject(ctx context.Context, tenantID, name string) (Project, error)
	GetProject(ctx context.Context, tenantID, projectID string) (Project, error)
	DeleteProject(ctx context.Context, tenantID, projectID string) error
}

type SQLiteProjectRepo struct {
	db *sqlx.DB
}

func NewSQLiteProjectRepo(db *sqlx.DB) *SQLiteProjectRepo {
	return &SQLiteProjectRepo{db: db}
}

func (r *SQLiteProjectRepo) ListProjects(ctx context.Context, tenantID string) ([]Project, error) {
	var projects []Project

	query := `SELECT id, name, created_at
				FROM projects
				WHERE tenant_id = ?
				ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &projects, query, tenantID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *SQLiteProjectRepo) CreateProject(ctx context.Context, tenantID, name string) (Project, error) {
	project := Project{
		Id:        uuid.New().String(),
		Name:      name,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
	}

	query := `INSERT INTO projects (id, name, tenant_id, created_at)
			  VALUES (?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query, project.Id, project.Name, tenantID, project.CreatedAt)
	if err != nil {
		return Project{}, err
	}

	return project, nil
}

func (r *SQLiteProjectRepo) GetProject(ctx context.Context, tenantID, projectID string) (Project, error) {
	var project Project

	query := `Select id, name, created_at 
				FROM projects 
				WHERE tenant_id = ? AND id = ?`

	err := r.db.GetContext(ctx, &project, query, tenantID, projectID)
	if err != nil {
		return Project{}, err
	}

	return project, nil
}

func (r *SQLiteProjectRepo) DeleteProject(ctx context.Context, tenantID, projectID string) error {
	query := `DELETE FROM projects WHERE tenant_id = ? AND id = ?`

	_, err := r.db.ExecContext(ctx, query, tenantID, projectID)
	if err != nil {
		return err
	}
	return nil
}
