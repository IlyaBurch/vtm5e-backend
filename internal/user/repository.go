package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, email, passwordHash string, username *string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, username)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, username, avatar_url, created_at`,
		email, passwordHash, username,
	).Scan(&u.ID, &u.Email, &u.Username, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (r *Repository) CreateOAuth(ctx context.Context, email, googleID string, username *string, avatarURL *string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, google_id, username, avatar_url)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, username, avatar_url, created_at`,
		email, googleID, username, avatarURL,
	).Scan(&u.ID, &u.Email, &u.Username, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, username, avatar_url, google_id, password_hash, created_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.Username, &u.AvatarURL, &u.GoogleID, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, username, avatar_url, created_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.Username, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, username, avatar_url, google_id, password_hash, created_at
		 FROM users WHERE google_id = $1`,
		googleID,
	).Scan(&u.ID, &u.Email, &u.Username, &u.AvatarURL, &u.GoogleID, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by google id: %w", err)
	}
	return u, nil
}

func (r *Repository) SetGoogleID(ctx context.Context, userID, googleID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET google_id = $1 WHERE id = $2`,
		googleID, userID,
	)
	if err != nil {
		return fmt.Errorf("set google id: %w", err)
	}
	return nil
}

func (r *Repository) UpdateProfile(ctx context.Context, userID string, username *string, avatarURL *string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx,
		`UPDATE users
		 SET username   = COALESCE($2, username),
		     avatar_url = COALESCE($3, avatar_url)
		 WHERE id = $1
		 RETURNING id, email, username, avatar_url, created_at`,
		userID, username, avatarURL,
	).Scan(&u.ID, &u.Email, &u.Username, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return u, nil
}

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`,
		username,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check username exists: %w", err)
	}
	return exists, nil
}
