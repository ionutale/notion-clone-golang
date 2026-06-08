package workspace

import (
	"context"
	"time"
)

type MockWorkspaceRepo struct {
	CreateFn     func(ctx context.Context, name, ownerID string) (*Workspace, error)
	ListByUserFn func(ctx context.Context, userID string) ([]Workspace, error)
	GetByIDFn    func(ctx context.Context, id string) (*Workspace, error)
	UpdateFn     func(ctx context.Context, id, name string) error
	DeleteFn     func(ctx context.Context, id string) error
	IsMemberFn   func(ctx context.Context, workspaceID, userID string) (bool, error)
	AddMemberFn  func(ctx context.Context, workspaceID, userID, role string) error
	RemoveMemberFn func(ctx context.Context, workspaceID, userID string) error
}

func (m *MockWorkspaceRepo) Create(ctx context.Context, name, ownerID string) (*Workspace, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, name, ownerID)
	}
	return &Workspace{ID: "mock-ws", Name: name, OwnerID: ownerID, CreatedAt: time.Now()}, nil
}

func (m *MockWorkspaceRepo) ListByUser(ctx context.Context, userID string) ([]Workspace, error) {
	if m.ListByUserFn != nil {
		return m.ListByUserFn(ctx, userID)
	}
	return []Workspace{}, nil
}

func (m *MockWorkspaceRepo) GetByID(ctx context.Context, id string) (*Workspace, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return &Workspace{ID: id, Name: "Test Workspace", OwnerID: "owner-id", CreatedAt: time.Now()}, nil
}

func (m *MockWorkspaceRepo) Update(ctx context.Context, id, name string) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, id, name)
	}
	return nil
}

func (m *MockWorkspaceRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *MockWorkspaceRepo) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	if m.IsMemberFn != nil {
		return m.IsMemberFn(ctx, workspaceID, userID)
	}
	return true, nil
}

func (m *MockWorkspaceRepo) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	if m.AddMemberFn != nil {
		return m.AddMemberFn(ctx, workspaceID, userID, role)
	}
	return nil
}

func (m *MockWorkspaceRepo) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	if m.RemoveMemberFn != nil {
		return m.RemoveMemberFn(ctx, workspaceID, userID)
	}
	return nil
}
