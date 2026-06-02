package workspace

import "time"

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

type CreateRequest struct {
	Name string `json:"name"`
}

type InviteRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}
