package workspace

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("workspace not found")
	ErrNotOwner = errors.New("only the owner can perform this action")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
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
	ok, err := s.repo.IsMember(ctx, id, userID)
	if err != nil || !ok {
		return ErrNotFound
	}
	return s.repo.Update(ctx, id, name)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
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

func (s *Service) InviteMember(ctx context.Context, workspaceID, email, role, requesterID string) error {
	w, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return ErrNotFound
	}
	if w.OwnerID != requesterID {
		return ErrNotOwner
	}
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, workspaceID, memberID, requesterID string) error {
	w, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return ErrNotFound
	}
	if w.OwnerID != requesterID {
		return ErrNotOwner
	}
	return s.repo.RemoveMember(ctx, workspaceID, memberID)
}
