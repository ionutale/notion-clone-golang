package block

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo *Repository
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		repo: NewRepository(pool),
	}
}

func (s *Service) CreatePage(ctx context.Context, workspaceID, userID uuid.UUID, title string) (Block, error) {
	content, _ := json.Marshal(map[string]string{"title": title})
	page := Block{
		WorkspaceID: workspaceID,
		Type:        TypePage,
		Content:     content,
		CreatedBy:   &userID,
		Position:    1 << 31,
	}

	initialContent, _ := json.Marshal(map[string]string{"html": ""})
	initial := Block{
		WorkspaceID: workspaceID,
		Type:        TypeText,
		Content:     initialContent,
		CreatedBy:   &userID,
		Position:    1 << 31,
	}

	if err := s.repo.CreatePageWithInitial(ctx, &page, &initial); err != nil {
		return Block{}, fmt.Errorf("create page: %w", err)
	}
	return page, nil
}

func (s *Service) GetPageTree(ctx context.Context, pageID uuid.UUID) (PageTree, error) {
	page, err := s.repo.GetByID(ctx, pageID)
	if err != nil {
		return PageTree{}, fmt.Errorf("get page: %w", err)
	}
	if page.Type != TypePage {
		return PageTree{}, fmt.Errorf("not a page")
	}
	blocks, err := s.repo.GetTree(ctx, pageID)
	if err != nil {
		return PageTree{}, fmt.Errorf("get tree: %w", err)
	}
	return PageTree{Page: page, Blocks: blocks}, nil
}

func (s *Service) ListPages(ctx context.Context, workspaceID uuid.UUID, cursor *int64, limit int) ([]PageSummary, *int64, error) {
	return s.repo.ListPages(ctx, workspaceID, cursor, limit)
}

func (s *Service) CreateBlock(ctx context.Context, workspaceID, userID uuid.UUID, req CreateBlockRequest) (Block, error) {
	if !ValidTypes[req.Type] {
		return Block{}, fmt.Errorf("invalid block type: %s", req.Type)
	}

	content := req.Content
	if content == nil {
		content = json.RawMessage("{}")
	}

	parentID := req.ParentID
	var position int64

	if parentID != nil {
		siblings, err := s.repo.GetSiblings(ctx, parentID, workspaceID)
		if err == nil && len(siblings) > 0 {
			last := siblings[len(siblings)-1]
			after := last.Position + (1 << 31)
			position = after
		} else {
			position = 1 << 31
		}
	} else {
		position = 1 << 31
	}

	if req.Position != nil {
		position = *req.Position
	}

	block := Block{
		WorkspaceID: workspaceID,
		ParentID:    parentID,
		Type:        req.Type,
		Content:     content,
		Position:    position,
		CreatedBy:   &userID,
	}
	if err := s.repo.Create(ctx, &block); err != nil {
		return Block{}, fmt.Errorf("create block: %w", err)
	}
	return block, nil
}

func (s *Service) UpdateBlock(ctx context.Context, id uuid.UUID, req UpdateBlockRequest) (Block, error) {
	if req.Type != nil && !ValidTypes[*req.Type] {
		return Block{}, fmt.Errorf("invalid block type: %s", *req.Type)
	}
	block, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return Block{}, fmt.Errorf("update block: %w", err)
	}
	return block, nil
}

func (s *Service) DeleteBlock(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) RestoreBlock(ctx context.Context, id uuid.UUID) (Block, error) {
	return s.repo.Restore(ctx, id)
}

func (s *Service) MoveBlock(ctx context.Context, workspaceID uuid.UUID, id uuid.UUID, req MoveBlockRequest) (Block, error) {
	siblings, err := s.repo.GetSiblings(ctx, req.ParentID, workspaceID)
	if err != nil {
		return Block{}, fmt.Errorf("get siblings: %w", err)
	}

	var before, after *int64
	for _, sib := range siblings {
		if sib.ID == id {
			continue
		}
		if sib.Position < req.Position && (before == nil || sib.Position > *before) {
			before = &sib.Position
		}
		if sib.Position > req.Position && (after == nil || sib.Position < *after) {
			after = &sib.Position
		}
	}

	position := MiddlePosition(before, after)
	moveReq := MoveBlockRequest{
		ParentID: req.ParentID,
		Position: position,
	}

	return s.repo.Move(ctx, id, moveReq)
}

func (s *Service) Search(ctx context.Context, workspaceID uuid.UUID, query string, limit, offset int) ([]SearchResult, error) {
	return s.repo.Search(ctx, workspaceID, query, limit, offset)
}

func (s *Service) StartCleanupLoop(ctx context.Context, interval time.Duration, days int) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := s.repo.CleanupExpired(ctx, nil, days); err != nil {
					slog.Error("background cleanup failed", "error", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Service) ListFavorites(ctx context.Context, workspaceID uuid.UUID, cursor *int64, limit int) ([]PageSummary, *int64, error) {
	return s.repo.ListFavorites(ctx, workspaceID, cursor, limit)
}

func (s *Service) ListTrash(ctx context.Context, workspaceID uuid.UUID, cursor *time.Time, limit int) ([]PageSummary, *time.Time, error) {
	return s.repo.ListTrash(ctx, workspaceID, cursor, limit)
}

func (s *Service) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	return s.repo.PermanentDelete(ctx, id)
}

func (s *Service) SplitBlock(ctx context.Context, workspaceID, userID uuid.UUID, id uuid.UUID, splitPosition int) (Block, Block, error) {
	original, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Block{}, Block{}, fmt.Errorf("get block: %w", err)
	}

	var content map[string]interface{}
	if err := json.Unmarshal(original.Content, &content); err != nil {
		return Block{}, Block{}, fmt.Errorf("parse content: %w", err)
	}

	richText, ok := content["rich_text"].([]interface{})
	if !ok {
		return Block{}, Block{}, fmt.Errorf("no rich_text in content")
	}

	leftText := append([]interface{}{}, richText[:splitPosition]...)
	rightText := append([]interface{}{}, richText[splitPosition:]...)

	leftContent, err := json.Marshal(mergeMaps(content, map[string]interface{}{"rich_text": leftText}))
	if err != nil {
		return Block{}, Block{}, fmt.Errorf("marshal left content: %w", err)
	}

	rightContent, err := json.Marshal(mergeMaps(content, map[string]interface{}{"rich_text": rightText}))
	if err != nil {
		return Block{}, Block{}, fmt.Errorf("marshal right content: %w", err)
	}

	newBlock := Block{
		WorkspaceID: workspaceID,
		ParentID:    original.ParentID,
		Type:        original.Type,
		Content:     rightContent,
		CreatedBy:   &userID,
	}

	updated, err := s.repo.SplitBlockTx(ctx, id, leftContent, &newBlock)
	if err != nil {
		return Block{}, Block{}, fmt.Errorf("split block: %w", err)
	}

	return updated, newBlock, nil
}

func mergeMaps(base, overlay map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(base))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overlay {
		result[k] = v
	}
	return result
}

func (s *Service) MergeBlocks(ctx context.Context, sourceID, targetID uuid.UUID) (Block, error) {
	source, err := s.repo.GetByID(ctx, sourceID)
	if err != nil {
		return Block{}, fmt.Errorf("get source: %w", err)
	}
	target, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return Block{}, fmt.Errorf("get target: %w", err)
	}

	var sc, tc map[string]interface{}
	if err := json.Unmarshal(source.Content, &sc); err != nil {
		return Block{}, fmt.Errorf("parse source content: %w", err)
	}
	if err := json.Unmarshal(target.Content, &tc); err != nil {
		return Block{}, fmt.Errorf("parse target content: %w", err)
	}

	srcRich, ok := sc["rich_text"].([]interface{})
	if !ok {
		srcRich = []interface{}{}
	}
	tgtRich, ok := tc["rich_text"].([]interface{})
	if !ok {
		tgtRich = []interface{}{}
	}

	merged := append(tgtRich, srcRich...)
	tc["rich_text"] = merged
	mergedContent, err := json.Marshal(tc)
	if err != nil {
		return Block{}, fmt.Errorf("marshal merged content: %w", err)
	}

	updated, err := s.repo.MergeBlocksTx(ctx, targetID, mergedContent, sourceID)
	if err != nil {
		return Block{}, fmt.Errorf("merge blocks: %w", err)
	}

	return updated, nil
}
