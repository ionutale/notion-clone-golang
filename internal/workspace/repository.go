package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, name, ownerID string) (*Workspace, error) {
	w := &Workspace{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO workspaces (name, owner_id) VALUES ($1, $2)
		 RETURNING id, name, owner_id, created_at`,
		name, ownerID,
	).Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'owner')`,
		w.ID, ownerID,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Workspace, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT w.id, w.name, w.owner_id, w.created_at
		 FROM workspaces w
		 JOIN workspace_members wm ON w.id = wm.workspace_id
		 WHERE wm.user_id = $1
		 ORDER BY w.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var workspaces []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, w)
	}
	return workspaces, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Workspace, error) {
	w := &Workspace{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, owner_id, created_at FROM workspaces WHERE id = $1`,
		id,
	).Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *Repository) Update(ctx context.Context, id, name string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE workspaces SET name = $1 WHERE id = $2`,
		name, id,
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id)
	return err
}

func (r *Repository) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	).Scan(&count)
	return count > 0, err
}

func (r *Repository) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT DO NOTHING`,
		workspaceID, userID, role,
	)
	return err
}

func (r *Repository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	)
	return err
}
