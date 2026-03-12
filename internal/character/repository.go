package character

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]ListItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, clan, updated_at FROM characters WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list characters: %w", err)
	}
	defer rows.Close()

	var items []ListItem
	for rows.Next() {
		var item ListItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Clan, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan character list item: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []ListItem{}
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (*Character, error) {
	c := &Character{}
	var dataBytes []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, clan, data, created_at, updated_at FROM characters WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Clan, &dataBytes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get character: %w", err)
	}
	c.Data = json.RawMessage(dataBytes)
	return c, nil
}

func (r *Repository) Create(ctx context.Context, userID, name, clan string, data json.RawMessage) (*Character, error) {
	c := &Character{}
	var dataBytes []byte
	err := r.db.QueryRow(ctx,
		`INSERT INTO characters (user_id, name, clan, data) VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, name, clan, data, created_at, updated_at`,
		userID, name, clan, []byte(data),
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Clan, &dataBytes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create character: %w", err)
	}
	c.Data = json.RawMessage(dataBytes)
	return c, nil
}

func (r *Repository) Update(ctx context.Context, id int, name, clan string, data json.RawMessage) (*Character, error) {
	c := &Character{}
	var dataBytes []byte
	err := r.db.QueryRow(ctx,
		`UPDATE characters SET name = $2, clan = $3, data = $4, updated_at = now()
		 WHERE id = $1
		 RETURNING id, user_id, name, clan, data, created_at, updated_at`,
		id, name, clan, []byte(data),
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Clan, &dataBytes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update character: %w", err)
	}
	c.Data = json.RawMessage(dataBytes)
	return c, nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM characters WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete character: %w", err)
	}
	return nil
}
