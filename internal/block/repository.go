package block

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, b *Block) error {
	b.ID = uuid.New()
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now

	content := b.Content
	if content == nil {
		content = json.RawMessage("{}")
	}

	var path *string
	if b.ParentID != nil {
		var parentPath string
		err := r.pool.QueryRow(ctx, `SELECT path::text FROM blocks WHERE id = $1`, *b.ParentID).Scan(&parentPath)
		if err == nil {
			p := parentPath + "." + b.ID.String()
			path = &p
		}
	} else {
		p := b.ID.String()
		path = &p
	}
	b.Path = path

	if b.Position == 0 {
		var maxPos *int64
		err := r.pool.QueryRow(ctx,
			`SELECT MAX(position) FROM blocks WHERE parent_id IS NOT DISTINCT FROM $1 AND deleted_at IS NULL`,
			b.ParentID).Scan(&maxPos)
		if err == nil && maxPos != nil {
			b.Position = *maxPos + (1 << 31)
		} else {
			b.Position = 1 << 31
		}
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO blocks (id, workspace_id, parent_id, type, content, position, path, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::ltree, $8, $9, $10)`,
		b.ID, b.WorkspaceID, b.ParentID, b.Type, content, b.Position, *b.Path, b.CreatedBy, b.CreatedAt, b.UpdatedAt)
	return err
}

func scanBlock(row pgx.Row) (Block, error) {
	var b Block
	err := row.Scan(
		&b.ID, &b.WorkspaceID, &b.ParentID, &b.Type, &b.Content,
		&b.Position, &b.Path, &b.CreatedBy, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
	)
	return b, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Block, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, workspace_id, parent_id, type, content, position, path::text, created_by, created_at, updated_at, deleted_at
		FROM blocks WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanBlock(row)
}

func (r *Repository) GetTree(ctx context.Context, pageID uuid.UUID) ([]Block, error) {
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE block_tree AS (
			SELECT id, workspace_id, parent_id, type, content, position, path::text, created_by, created_at, updated_at, deleted_at
			FROM blocks WHERE id = $1 AND deleted_at IS NULL
			UNION ALL
			SELECT b.id, b.workspace_id, b.parent_id, b.type, b.content, b.position, b.path::text, b.created_by, b.created_at, b.updated_at, b.deleted_at
			FROM blocks b
			INNER JOIN block_tree bt ON b.parent_id = bt.id
			WHERE b.deleted_at IS NULL
		)
		SELECT * FROM block_tree ORDER BY path, position`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		b, err := scanBlock(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (r *Repository) ListPages(ctx context.Context, workspaceID uuid.UUID) ([]PageSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, content->>'title' AS title, content->>'icon' AS icon, content->>'icon_type' AS icon_type, created_at, updated_at
		FROM blocks
		WHERE workspace_id = $1 AND type = 'page' AND parent_id IS NULL AND deleted_at IS NULL
		ORDER BY position`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []PageSummary
	for rows.Next() {
		var p PageSummary
		if err := rows.Scan(&p.ID, &p.Title, &p.Icon, &p.IconType, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	return pages, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, req UpdateBlockRequest) (Block, error) {
	if req.Content != nil {
		_, err := r.pool.Exec(ctx, `UPDATE blocks SET content = $1, updated_at = now() WHERE id = $2 AND deleted_at IS NULL`, req.Content, id)
		if err != nil {
			return Block{}, err
		}
	}
	if req.Type != nil {
		_, err := r.pool.Exec(ctx, `UPDATE blocks SET type = $1, updated_at = now() WHERE id = $2 AND deleted_at IS NULL`, *req.Type, id)
		if err != nil {
			return Block{}, err
		}
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE blocks SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

func (r *Repository) Restore(ctx context.Context, id uuid.UUID) (Block, error) {
	_, err := r.pool.Exec(ctx, `UPDATE blocks SET deleted_at = NULL, updated_at = now() WHERE id = $1`, id)
	if err != nil {
		return Block{}, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Move(ctx context.Context, id uuid.UUID, req MoveBlockRequest) (Block, error) {
	_, err := r.pool.Exec(ctx, `UPDATE blocks SET parent_id = $1, position = $2, updated_at = now() WHERE id = $3 AND deleted_at IS NULL`,
		req.ParentID, req.Position, id)
	if err != nil {
		return Block{}, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) UpdatePath(ctx context.Context, id uuid.UUID, path string) error {
	_, err := r.pool.Exec(ctx, `UPDATE blocks SET path = $1::ltree, updated_at = now() WHERE id = $2`, path, id)
	return err
}

func MiddlePosition(before, after *int64) int64 {
	if before == nil && after == nil {
		return 1 << 31
	}
	if before == nil {
		return *after / 2
	}
	if after == nil {
		return *before + (1 << 31)
	}
	return (*before + *after) / 2
}

func (r *Repository) GetSiblings(ctx context.Context, parentID *uuid.UUID, workspaceID uuid.UUID) ([]Block, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workspace_id, parent_id, type, content, position, path::text, created_by, created_at, updated_at, deleted_at
		FROM blocks
		WHERE parent_id IS NOT DISTINCT FROM $1 AND workspace_id = $2 AND deleted_at IS NULL
		ORDER BY position`, parentID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		b, err := scanBlock(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (r *Repository) InsertAtPosition(ctx context.Context, b *Block, position int64) error {
	b.Position = position
	return r.Create(ctx, b)
}
