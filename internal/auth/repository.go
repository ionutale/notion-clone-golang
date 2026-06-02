package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
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

func hashToken(bytes []byte) string {
	sum := sha256.Sum256(bytes)
	return hex.EncodeToString(sum[:])
}

func (r *Repository) CreateRefreshToken(ctx context.Context, userID string, expiresAt string) (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	tokenHash := hashToken(bytes)
	_, err = r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (r *Repository) findAndConsumeRefreshToken(ctx context.Context, tokenHex string, consume bool) (string, string, error) {
	bytes, err := hex.DecodeString(tokenHex)
	if err != nil {
		return "", "", err
	}
	tokenHash := hashToken(bytes)
	var id, userID, storedHash string
	err = r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash FROM refresh_tokens
		 WHERE token_hash = $1 AND expires_at > now()`,
		tokenHash,
	).Scan(&id, &userID, &storedHash)
	if err != nil {
		return "", "", err
	}
	if consume {
		_, err = r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, id)
		if err != nil {
			log.Printf("failed to delete consumed refresh token %s: %v", id, err)
		}
	}
	return id, userID, nil
}

func (r *Repository) ValidateRefreshToken(ctx context.Context, tokenHex string) (string, error) {
	_, userID, err := r.findAndConsumeRefreshToken(ctx, tokenHex, true)
	if err != nil {
		return "", err
	}
	return userID, nil
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
	tokenHash := hashToken(bytes)
	user := &User{}
	var id, userID string
	err = r.pool.QueryRow(ctx,
		`SELECT rt.id, rt.user_id, u.id, u.email, u.name, u.created_at
		 FROM refresh_tokens rt JOIN users u ON rt.user_id = u.id
		 WHERE rt.token_hash = $1 AND rt.expires_at > now()`,
		tokenHash,
	).Scan(&id, &userID, &user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	_, err = r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, id)
	if err != nil {
		log.Printf("failed to delete consumed refresh token %s: %v", id, err)
	}
	return user, nil
}
