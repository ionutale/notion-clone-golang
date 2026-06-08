package workspace

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("workspace not found")
	ErrNotOwner = errors.New("only the owner can perform this action")
)

type WorkspaceRepository interface {
	Create(ctx context.Context, name, ownerID string) (*Workspace, error)
	ListByUser(ctx context.Context, userID string) ([]Workspace, error)
	GetByID(ctx context.Context, id string) (*Workspace, error)
	Update(ctx context.Context, id, name string) error
	Delete(ctx context.Context, id string) error
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
	AddMember(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
	ListMembers(ctx context.Context, workspaceID string) ([]MemberWithUser, error)
}

type Service struct {
	repo WorkspaceRepository
}

func NewService(repo WorkspaceRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, ownerID string) (*Workspace, error) {
	return s.repo.Create(ctx, name, ownerID)
}

func (s *Service) List(ctx context.Context, userID string) ([]Workspace, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, id string) (*Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id, name, userID string) error {
	w, err := s.repo.GetByID(ctx, id)
	if w == nil || err != nil {
		return ErrNotFound
	}
	if w.OwnerID != userID {
		return ErrNotOwner
	}
	return s.repo.Update(ctx, id, name)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	w, err := s.repo.GetByID(ctx, id)
	if w == nil || err != nil {
		return ErrNotFound
	}
	if w.OwnerID != userID {
		return ErrNotOwner
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	return s.repo.IsMember(ctx, workspaceID, userID)
}

func (s *Service) InviteMember(ctx context.Context, workspaceID, memberID, role, requesterID string) error {
	w, err := s.repo.GetByID(ctx, workspaceID)
	if w == nil || err != nil {
		return ErrNotFound
	}
	if w.OwnerID != requesterID {
		return ErrNotOwner
	}
	return s.repo.AddMember(ctx, workspaceID, memberID, role)
}

func (s *Service) ListMembers(ctx context.Context, workspaceID string) ([]MemberWithUser, error) {
	return s.repo.ListMembers(ctx, workspaceID)
}

func (s *Service) RemoveMember(ctx context.Context, workspaceID, memberID, requesterID string) error {
	w, err := s.repo.GetByID(ctx, workspaceID)
	if w == nil || err != nil {
		return ErrNotFound
	}
	if w.OwnerID != requesterID {
		return ErrNotOwner
	}
	return s.repo.RemoveMember(ctx, workspaceID, memberID)
}
