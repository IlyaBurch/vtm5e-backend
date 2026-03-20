package morkborg

import (
	"encoding/json"
	"time"
)

type Character struct {
	ID          string          `json:"id"`
	UserID      string          `json:"userId"`
	Name        string          `json:"name"`
	ClassName   string          `json:"className"`
	Description string          `json:"description"`
	ColdBlood   string          `json:"coldBlood"`
	Exhaustion  string          `json:"exhaustion"`
	Abilities   json.RawMessage `json:"abilities"`
	Strength    int             `json:"strength"`
	Agility     int             `json:"agility"`
	Toughness   int             `json:"toughness"`
	Presence    int             `json:"presence"`
	HP          int             `json:"hp"`
	MaxHP       int             `json:"maxHp"`
	Weapons     json.RawMessage `json:"weapons"`
	ArmorName   string          `json:"armorName"`
	ArmorTier   string          `json:"armorTier"`
	Sufferings  json.RawMessage `json:"sufferings"`
	Equipment   json.RawMessage `json:"equipment"`
	Silver      int             `json:"silver"`
	Omens       int             `json:"omens"`
	OmensUsed   int             `json:"omensUsed"`
	Notes       string          `json:"notes"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type CreateRequest struct {
	Name        string          `json:"name"`
	ClassName   string          `json:"className"`
	Description string          `json:"description"`
	ColdBlood   string          `json:"coldBlood"`
	Exhaustion  string          `json:"exhaustion"`
	Abilities   json.RawMessage `json:"abilities"`
	Strength    int             `json:"strength"`
	Agility     int             `json:"agility"`
	Toughness   int             `json:"toughness"`
	Presence    int             `json:"presence"`
	HP          int             `json:"hp"`
	MaxHP       int             `json:"maxHp"`
	Weapons     json.RawMessage `json:"weapons"`
	ArmorName   string          `json:"armorName"`
	ArmorTier   string          `json:"armorTier"`
	Sufferings  json.RawMessage `json:"sufferings"`
	Equipment   json.RawMessage `json:"equipment"`
	Silver      int             `json:"silver"`
	Omens       int             `json:"omens"`
	OmensUsed   int             `json:"omensUsed"`
	Notes       string          `json:"notes"`
}

type PatchRequest struct {
	Name        *string          `json:"name"`
	ClassName   *string          `json:"className"`
	Description *string          `json:"description"`
	ColdBlood   *string          `json:"coldBlood"`
	Exhaustion  *string          `json:"exhaustion"`
	Abilities   *json.RawMessage `json:"abilities"`
	Strength    *int             `json:"strength"`
	Agility     *int             `json:"agility"`
	Toughness   *int             `json:"toughness"`
	Presence    *int             `json:"presence"`
	HP          *int             `json:"hp"`
	MaxHP       *int             `json:"maxHp"`
	Weapons     *json.RawMessage `json:"weapons"`
	ArmorName   *string          `json:"armorName"`
	ArmorTier   *string          `json:"armorTier"`
	Sufferings  *json.RawMessage `json:"sufferings"`
	Equipment   *json.RawMessage `json:"equipment"`
	Silver      *int             `json:"silver"`
	Omens       *int             `json:"omens"`
	OmensUsed   *int             `json:"omensUsed"`
	Notes       *string          `json:"notes"`
}
