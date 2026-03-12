package character

import (
	"encoding/json"
	"time"
)

type Character struct {
	ID        int             `json:"id"`
	UserID    string          `json:"user_id"`
	Name      string          `json:"name"`
	Clan      string          `json:"clan"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type ListItem struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Clan      string    `json:"clan"`
	UpdatedAt time.Time `json:"updated_at"`
}
