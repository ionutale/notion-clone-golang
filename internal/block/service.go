package block

import (
	"context"
	"encoding/json"
	"fmt"

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
	block := Block{
		WorkspaceID: workspaceID,
		Type:        TypePage,
		Content:     content,
		CreatedBy:   &userID,
	}
	if err := s.repo.Create(ctx, &block); err != nil {
		return Block{}, fmt.Errorf("create page: %w", err)
	}
	return block, nil
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

func (s *Service) ListPages(ctx context.Context, workspaceID uuid.UUID) ([]PageSummary, error) {
	return s.repo.ListPages(ctx, workspaceID)
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

func (s *Service) ListFavorites(ctx context.Context, workspaceID uuid.UUID) ([]PageSummary, error) {
	return s.repo.ListFavorites(ctx, workspaceID)
}

func (s *Service) ListTrash(ctx context.Context, workspaceID uuid.UUID) ([]PageSummary, error) {
	_ = s.repo.CleanupExpired(ctx, workspaceID, 30)
	return s.repo.ListTrash(ctx, workspaceID)
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

	leftText := richText[:splitPosition]
	rightText := richText[splitPosition:]

	content["rich_text"] = leftText
	leftContent, _ := json.Marshal(content)

	content["rich_text"] = rightText
	rightContent, _ := json.Marshal(content)

	updated, err := s.repo.Update(ctx, id, UpdateBlockRequest{Content: leftContent})
	if err != nil {
		return Block{}, Block{}, fmt.Errorf("update original: %w", err)
	}

	newBlock := Block{
		WorkspaceID: workspaceID,
		ParentID:    original.ParentID,
		Type:        original.Type,
		Content:     rightContent,
		CreatedBy:   &userID,
	}
	if err := s.repo.Create(ctx, &newBlock); err != nil {
		return Block{}, Block{}, fmt.Errorf("create new block: %w", err)
	}

	_, _ = s.repo.Update(ctx, newBlock.ID, UpdateBlockRequest{Content: rightContent})

	return updated, newBlock, nil
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
	json.Unmarshal(source.Content, &sc)
	json.Unmarshal(target.Content, &tc)

	srcRich, _ := sc["rich_text"].([]interface{})
	tgtRich, _ := tc["rich_text"].([]interface{})

	merged := append(tgtRich, srcRich...)
	tc["rich_text"] = merged
	mergedContent, _ := json.Marshal(tc)

	updated, err := s.repo.Update(ctx, targetID, UpdateBlockRequest{Content: mergedContent})
	if err != nil {
		return Block{}, fmt.Errorf("merge content: %w", err)
	}

	if err := s.repo.SoftDelete(ctx, sourceID); err != nil {
		return Block{}, fmt.Errorf("delete source: %w", err)
	}

	return updated, nil
}
