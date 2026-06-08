package auth

import (
	"context"
	"time"
)

type MockUserRepo struct {
	CreateUserFn              func(ctx context.Context, email, password, name string) (*User, error)
	GetUserByEmailFn          func(ctx context.Context, email string) (*User, error)
	GetUserByIDFn             func(ctx context.Context, id string) (*User, error)
	CreateRefreshTokenFn      func(ctx context.Context, userID, expiresAt string) (string, error)
	GetUserByRefreshTokenFn   func(ctx context.Context, tokenHex string) (*User, error)
	DeleteUserRefreshTokensFn func(ctx context.Context, userID string) error
	UpdateUserFn              func(ctx context.Context, id, name, email string) error
	UpdatePasswordFn          func(ctx context.Context, id, password string) error
	DeleteUserFn              func(ctx context.Context, id string) error
}

func (m *MockUserRepo) CreateUser(ctx context.Context, email, password, name string) (*User, error) {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, email, password, name)
	}
	return &User{ID: "mock-id", Email: email, Name: name, CreatedAt: time.Now()}, nil
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if m.GetUserByEmailFn != nil {
		return m.GetUserByEmailFn(ctx, email)
	}
	return nil, ErrInvalidCredentials
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, id string) (*User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, id)
	}
	return &User{ID: id, Email: "test@test.com", Name: "Test", CreatedAt: time.Now()}, nil
}

func (m *MockUserRepo) CreateRefreshToken(ctx context.Context, userID, expiresAt string) (string, error) {
	if m.CreateRefreshTokenFn != nil {
		return m.CreateRefreshTokenFn(ctx, userID, expiresAt)
	}
	return "mock-refresh-token", nil
}

func (m *MockUserRepo) GetUserByRefreshToken(ctx context.Context, tokenHex string) (*User, error) {
	if m.GetUserByRefreshTokenFn != nil {
		return m.GetUserByRefreshTokenFn(ctx, tokenHex)
	}
	return nil, ErrInvalidCredentials
}

func (m *MockUserRepo) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	if m.DeleteUserRefreshTokensFn != nil {
		return m.DeleteUserRefreshTokensFn(ctx, userID)
	}
	return nil
}

func (m *MockUserRepo) UpdateUser(ctx context.Context, id, name, email string) error {
	if m.UpdateUserFn != nil {
		return m.UpdateUserFn(ctx, id, name, email)
	}
	return nil
}

func (m *MockUserRepo) UpdatePassword(ctx context.Context, id, password string) error {
	if m.UpdatePasswordFn != nil {
		return m.UpdatePasswordFn(ctx, id, password)
	}
	return nil
}

func (m *MockUserRepo) DeleteUser(ctx context.Context, id string) error {
	if m.DeleteUserFn != nil {
		return m.DeleteUserFn(ctx, id)
	}
	return nil
}

type MockWorkspaceCreator struct {
	CreateFn func(ctx context.Context, name, ownerID string) (interface{}, error)
}

func (m *MockWorkspaceCreator) Create(ctx context.Context, name, ownerID string) (interface{}, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, name, ownerID)
	}
	return nil, nil
}
