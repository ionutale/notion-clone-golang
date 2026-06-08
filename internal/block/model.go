package block

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type BlockType string

const (
	TypePage       BlockType = "page"
	TypeText       BlockType = "text"
	TypeHeading1   BlockType = "heading_1"
	TypeHeading2   BlockType = "heading_2"
	TypeHeading3   BlockType = "heading_3"
	TypeBulletList BlockType = "bullet_list_item"
	TypeNumberList BlockType = "numbered_list_item"
	TypeToggle     BlockType = "toggle"
	TypeDivider    BlockType = "divider"
	TypeImage      BlockType = "image"
)

var ValidTypes = map[BlockType]bool{
	TypePage: true, TypeText: true, TypeHeading1: true,
	TypeHeading2: true, TypeHeading3: true, TypeBulletList: true,
	TypeNumberList: true, TypeToggle: true, TypeDivider: true, TypeImage: true,
}

type Block struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id"`
	ParentID    *uuid.UUID      `json:"parent_id"`
	Type        BlockType        `json:"type"`
	Content     json.RawMessage `json:"content"`
	Position    int64            `json:"position"`
	Path        *string         `json:"path"`
	CreatedBy   *uuid.UUID      `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}

type CreateBlockRequest struct {
	ParentID *uuid.UUID `json:"parent_id"`
	Type     BlockType  `json:"type"`
	Content  json.RawMessage `json:"content,omitempty"`
	Position *int64     `json:"position,omitempty"`
}

type UpdateBlockRequest struct {
	Content json.RawMessage `json:"content,omitempty"`
	Type    *BlockType      `json:"type,omitempty"`
}

type MoveBlockRequest struct {
	ParentID *uuid.UUID `json:"parent_id"`
	Position int64      `json:"position"`
}

type PageSummary struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Icon      *string    `json:"icon,omitempty"`
	IconType  *string    `json:"icon_type,omitempty"`
	Position  int64      `json:"position"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type PageTree struct {
	Page   Block   `json:"page"`
	Blocks []Block `json:"blocks"`
}

type SearchResult struct {
	BlockID   uuid.UUID `json:"block_id"`
	PageID    uuid.UUID `json:"page_id"`
	PageTitle string    `json:"page_title"`
	BlockType string    `json:"block_type"`
	Excerpt   string    `json:"excerpt"`
	Rank      float64   `json:"rank"`
}
