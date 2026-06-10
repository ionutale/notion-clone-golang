package workspace

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, name, ownerID string) (*Workspace, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	w := &Workspace{}
	err = tx.QueryRow(ctx,
		`INSERT INTO workspaces (name, owner_id) VALUES ($1, $2)
		 RETURNING id, name, owner_id, created_at`,
		name, ownerID,
	).Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'owner')`,
		w.ID, ownerID,
	)
	if err != nil {
		return nil, err
	}
	return w, tx.Commit(ctx)
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
	if err := rows.Err(); err != nil {
		return nil, err
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
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
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM blocks WHERE workspace_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM workspace_members WHERE workspace_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
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

func (r *Repository) ListMembers(ctx context.Context, workspaceID string) ([]MemberWithUser, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT wm.user_id, u.email, u.name, wm.role, wm.joined_at
		 FROM workspace_members wm
		 JOIN users u ON wm.user_id = u.id
		 WHERE wm.workspace_id = $1
		 ORDER BY wm.joined_at ASC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []MemberWithUser
	for rows.Next() {
		var m MemberWithUser
		if err := rows.Scan(&m.UserID, &m.Email, &m.Name, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	)
	return err
}
