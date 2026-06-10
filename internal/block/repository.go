package block

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var path *string
	if b.ParentID != nil {
		var parentPath string
		err := tx.QueryRow(ctx, `SELECT path::text FROM blocks WHERE id = $1`, *b.ParentID).Scan(&parentPath)
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
		err := tx.QueryRow(ctx,
			`SELECT MAX(position) FROM blocks WHERE parent_id IS NOT DISTINCT FROM $1 AND deleted_at IS NULL FOR UPDATE`,
			b.ParentID).Scan(&maxPos)
		if err == nil && maxPos != nil {
			b.Position = *maxPos + (1 << 31)
		} else {
			b.Position = 1 << 31
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO blocks (id, workspace_id, parent_id, type, content, position, path, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::ltree, $8, $9, $10)`,
		b.ID, b.WorkspaceID, b.ParentID, b.Type, content, b.Position, *b.Path, b.CreatedBy, b.CreatedAt, b.UpdatedAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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

	blocks := make([]Block, 0)
	for rows.Next() {
		b, err := scanBlock(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (r *Repository) ListPages(ctx context.Context, workspaceID uuid.UUID, cursor *int64, limit int) ([]PageSummary, *int64, error) {
	query := `
		SELECT id, content->>'title' AS title, content->>'icon' AS icon, content->>'icon_type' AS icon_type, position, created_at, updated_at
		FROM blocks
		WHERE workspace_id = $1 AND type = 'page' AND parent_id IS NULL AND deleted_at IS NULL`
	args := []interface{}{workspaceID}
	if cursor != nil {
		args = append(args, *cursor)
		query += fmt.Sprintf(" AND position > $%d", len(args))
	}
	args = append(args, limit+1)
	query += fmt.Sprintf(" ORDER BY position LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	pages := make([]PageSummary, 0, limit+1)
	for rows.Next() {
		var p PageSummary
		if err := rows.Scan(&p.ID, &p.Title, &p.Icon, &p.IconType, &p.Position, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, nil, err
		}
		pages = append(pages, p)
	}

	var nextCursor *int64
	hasMore := len(pages) > limit
	if hasMore {
		pages = pages[:limit]
		last := pages[len(pages)-1]
		nextCursor = &last.Position
	}

	return pages, nextCursor, nil
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
	_, err := r.pool.Exec(ctx, `
		WITH RECURSIVE descendants AS (
			SELECT id FROM blocks WHERE id = $1 AND deleted_at IS NULL
			UNION ALL
			SELECT b.id FROM blocks b
			INNER JOIN descendants d ON b.parent_id = d.id
			WHERE b.deleted_at IS NULL
		)
		UPDATE blocks SET deleted_at = now(), updated_at = now()
		WHERE id IN (SELECT id FROM descendants)
	`, id)
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
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Block{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE blocks SET parent_id = $1, position = $2, updated_at = now() WHERE id = $3 AND deleted_at IS NULL`,
		req.ParentID, req.Position, id)
	if err != nil {
		return Block{}, err
	}

	// Rebuild path for the moved block
	if req.ParentID != nil {
		var parentPath string
		err = tx.QueryRow(ctx, `SELECT path::text FROM blocks WHERE id = $1`, *req.ParentID).Scan(&parentPath)
		if err == nil {
			newPath := parentPath + "." + id.String()
			_, err = tx.Exec(ctx, `UPDATE blocks SET path = $1::ltree WHERE id = $2`, newPath, id)
			if err != nil {
				return Block{}, err
			}
		}
	} else {
		_, err = tx.Exec(ctx, `UPDATE blocks SET path = $1::ltree WHERE id = $2`, id.String(), id)
		if err != nil {
			return Block{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Block{}, err
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) UpdatePath(ctx context.Context, id uuid.UUID, path string) error {
	_, err := r.pool.Exec(ctx, `UPDATE blocks SET path = $1::ltree, updated_at = now() WHERE id = $2`, path, id)
	return err
}

func (r *Repository) Search(ctx context.Context, workspaceID uuid.UUID, query string, limit, offset int) ([]SearchResult, error) {
	if query == "" {
		return make([]SearchResult, 0), nil
	}
	rows, err := r.pool.Query(ctx, `
		WITH RECURSIVE page_ancestors AS (
			SELECT b.id, b.id AS page_id, b.content->>'title' AS page_title
			FROM blocks b
			WHERE b.type = 'page' AND b.parent_id IS NULL AND b.deleted_at IS NULL

			UNION ALL

			SELECT b.id, pa.page_id, pa.page_title
			FROM blocks b
			JOIN page_ancestors pa ON b.parent_id = pa.id
			WHERE b.deleted_at IS NULL
		)
		SELECT b.id AS block_id, pa.page_id, pa.page_title, b.type AS block_type,
			left(regexp_replace(coalesce(b.content->>'html', b.content->>'title', ''), '<[^>]+>', '', 'g'), 150) AS excerpt,
			ts_rank(b.search_vector, plainto_tsquery('english', $2)) AS rank
		FROM blocks b
		JOIN page_ancestors pa ON b.id = pa.id
		WHERE b.workspace_id = $1
		  AND b.search_vector @@ plainto_tsquery('english', $2)
		  AND b.deleted_at IS NULL
		ORDER BY rank DESC
		LIMIT $3 OFFSET $4
	`, workspaceID, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]SearchResult, 0)
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.BlockID, &r.PageID, &r.PageTitle, &r.BlockType, &r.Excerpt, &r.Rank); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func (r *Repository) ListFavorites(ctx context.Context, workspaceID uuid.UUID, cursor *int64, limit int) ([]PageSummary, *int64, error) {
	query := `
		SELECT id, content->>'title' AS title, content->>'icon' AS icon, content->>'icon_type' AS icon_type, position, created_at, updated_at
		FROM blocks
		WHERE workspace_id = $1 AND type = 'page' AND parent_id IS NULL AND deleted_at IS NULL
		  AND (content->>'favorited')::boolean = true`
	args := []interface{}{workspaceID}
	if cursor != nil {
		args = append(args, *cursor)
		query += fmt.Sprintf(" AND position > $%d", len(args))
	}
	args = append(args, limit+1)
	query += fmt.Sprintf(" ORDER BY position LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	pages := make([]PageSummary, 0, limit+1)
	for rows.Next() {
		var p PageSummary
		if err := rows.Scan(&p.ID, &p.Title, &p.Icon, &p.IconType, &p.Position, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, nil, err
		}
		pages = append(pages, p)
	}

	var nextCursor *int64
	hasMore := len(pages) > limit
	if hasMore {
		pages = pages[:limit]
		nextCursor = &pages[len(pages)-1].Position
	}

	return pages, nextCursor, nil
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

	blocks := make([]Block, 0)
	for rows.Next() {
		b, err := scanBlock(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (r *Repository) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		WITH RECURSIVE descendants AS (
			SELECT id FROM blocks WHERE id = $1
			UNION ALL
			SELECT b.id FROM blocks b
			INNER JOIN descendants d ON b.parent_id = d.id
		)
		DELETE FROM blocks WHERE id IN (SELECT id FROM descendants)
	`, id)
	return err
}

func (r *Repository) ListTrash(ctx context.Context, workspaceID uuid.UUID, cursor *time.Time, limit int) ([]PageSummary, *time.Time, error) {
	query := `
		SELECT id, content->>'title' AS title, content->>'icon' AS icon, content->>'icon_type' AS icon_type, deleted_at, created_at, updated_at
		FROM blocks
		WHERE workspace_id = $1 AND type = 'page' AND parent_id IS NULL AND deleted_at IS NOT NULL`
	args := []interface{}{workspaceID}
	if cursor != nil {
		args = append(args, *cursor)
		query += fmt.Sprintf(" AND deleted_at < $%d", len(args))
	}
	args = append(args, limit+1)
	query += fmt.Sprintf(" ORDER BY deleted_at DESC LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	pages := make([]PageSummary, 0, limit+1)
	for rows.Next() {
		var p PageSummary
		if err := rows.Scan(&p.ID, &p.Title, &p.Icon, &p.IconType, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, nil, err
		}
		pages = append(pages, p)
	}

	var nextCursor *time.Time
	hasMore := len(pages) > limit
	if hasMore {
		pages = pages[:limit]
		nextCursor = pages[len(pages)-1].DeletedAt
	}

	return pages, nextCursor, nil
}

func (r *Repository) CleanupExpired(ctx context.Context, workspaceID *uuid.UUID, days int) error {
	var err error
	if workspaceID != nil {
		_, err = r.pool.Exec(ctx, `
			DELETE FROM blocks
			WHERE workspace_id = $1 AND deleted_at IS NOT NULL AND deleted_at < now() - ($2 || ' days')::interval
		`, *workspaceID, fmt.Sprintf("%d", days))
	} else {
		_, err = r.pool.Exec(ctx, `
			DELETE FROM blocks
			WHERE deleted_at IS NOT NULL AND deleted_at < now() - ($1 || ' days')::interval
		`, fmt.Sprintf("%d", days))
	}
	if err != nil {
		slog.Warn("cleanup expired failed", "error", err)
	}
	return nil
}

func (r *Repository) InsertAtPosition(ctx context.Context, b *Block, position int64) error {
	b.Position = position
	return r.Create(ctx, b)
}
