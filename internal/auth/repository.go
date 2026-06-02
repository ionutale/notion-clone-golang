package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, email, password, name string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &User{}
	err = r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
		 RETURNING id, email, name, created_at`,
		email, string(hash), name,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) CreateRefreshToken(ctx context.Context, userID string, expiresAt string) (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	hash, err := bcrypt.GenerateFromPassword(bytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, string(hash), expiresAt,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (r *Repository) ValidateRefreshToken(ctx context.Context, tokenHex string) (string, error) {
	bytes, err := hex.DecodeString(tokenHex)
	if err != nil {
		return "", err
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, token_hash FROM refresh_tokens WHERE expires_at > now()`,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var id, userID, hash string
		if err := rows.Scan(&id, &userID, &hash); err != nil {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), bytes) == nil {
			r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, id)
			return userID, nil
		}
	}
	return "", bcrypt.ErrMismatchedHashAndPassword
}

func (r *Repository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

func (r *Repository) GetUserByRefreshToken(ctx context.Context, tokenHex string) (*User, error) {
	bytes, err := hex.DecodeString(tokenHex)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx,
		`SELECT rt.id, rt.user_id, rt.token_hash, u.id, u.email, u.name, u.created_at
		 FROM refresh_tokens rt JOIN users u ON rt.user_id = u.id
		 WHERE rt.expires_at > now()`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, userID, hash string
		user := &User{}
		if err := rows.Scan(&id, &userID, &hash, &user.ID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), bytes) == nil {
			r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, id)
			return user, nil
		}
	}
	return nil, bcrypt.ErrMismatchedHashAndPassword
}
