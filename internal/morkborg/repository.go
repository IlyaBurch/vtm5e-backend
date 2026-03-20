package morkborg

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

const selectFields = `id, user_id, name, class_name, description, cold_blood, exhaustion,
	abilities, strength, agility, toughness, presence, hp, max_hp,
	weapons, armor_name, armor_tier, sufferings, equipment,
	silver, omens, omens_used, notes, created_at, updated_at`

func scanCharacter(row interface{ Scan(...any) error }) (*Character, error) {
	c := &Character{}
	err := row.Scan(
		&c.ID, &c.UserID, &c.Name, &c.ClassName, &c.Description, &c.ColdBlood, &c.Exhaustion,
		&c.Abilities, &c.Strength, &c.Agility, &c.Toughness, &c.Presence, &c.HP, &c.MaxHP,
		&c.Weapons, &c.ArmorName, &c.ArmorTier, &c.Sufferings, &c.Equipment,
		&c.Silver, &c.Omens, &c.OmensUsed, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]*Character, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+selectFields+` FROM morkborg_characters WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list morkborg characters: %w", err)
	}
	defer rows.Close()

	var chars []*Character
	for rows.Next() {
		c, err := scanCharacter(rows)
		if err != nil {
			return nil, fmt.Errorf("scan morkborg character: %w", err)
		}
		chars = append(chars, c)
	}
	return chars, nil
}

func (r *Repository) Create(ctx context.Context, userID string, req *CreateRequest) (*Character, error) {
	abilities := req.Abilities
	if abilities == nil {
		abilities = []byte("[]")
	}
	weapons := req.Weapons
	if weapons == nil {
		weapons = []byte("[]")
	}
	sufferings := req.Sufferings
	if sufferings == nil {
		sufferings = []byte("[false,false,false,false,false,false]")
	}
	equipment := req.Equipment
	if equipment == nil {
		equipment = []byte("[]")
	}

	row := r.db.QueryRow(ctx,
		`INSERT INTO morkborg_characters
		 (user_id, name, class_name, description, cold_blood, exhaustion,
		  abilities, strength, agility, toughness, presence, hp, max_hp,
		  weapons, armor_name, armor_tier, sufferings, equipment,
		  silver, omens, omens_used, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		 RETURNING `+selectFields,
		userID, req.Name, req.ClassName, req.Description, req.ColdBlood, req.Exhaustion,
		abilities, req.Strength, req.Agility, req.Toughness, req.Presence, req.HP, req.MaxHP,
		weapons, req.ArmorName, req.ArmorTier, sufferings, equipment,
		req.Silver, req.Omens, req.OmensUsed, req.Notes,
	)

	c, err := scanCharacter(row)
	if err != nil {
		return nil, fmt.Errorf("create morkborg character: %w", err)
	}
	return c, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Character, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+selectFields+` FROM morkborg_characters WHERE id = $1`,
		id,
	)
	c, err := scanCharacter(row)
	if err != nil {
		return nil, fmt.Errorf("get morkborg character: %w", err)
	}
	return c, nil
}

func (r *Repository) Patch(ctx context.Context, id string, req *PatchRequest) (*Character, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE morkborg_characters SET
		  name        = COALESCE($2,  name),
		  class_name  = COALESCE($3,  class_name),
		  description = COALESCE($4,  description),
		  cold_blood  = COALESCE($5,  cold_blood),
		  exhaustion  = COALESCE($6,  exhaustion),
		  abilities   = COALESCE($7,  abilities),
		  strength    = COALESCE($8,  strength),
		  agility     = COALESCE($9,  agility),
		  toughness   = COALESCE($10, toughness),
		  presence    = COALESCE($11, presence),
		  hp          = COALESCE($12, hp),
		  max_hp      = COALESCE($13, max_hp),
		  weapons     = COALESCE($14, weapons),
		  armor_name  = COALESCE($15, armor_name),
		  armor_tier  = COALESCE($16, armor_tier),
		  sufferings  = COALESCE($17, sufferings),
		  equipment   = COALESCE($18, equipment),
		  silver      = COALESCE($19, silver),
		  omens       = COALESCE($20, omens),
		  omens_used  = COALESCE($21, omens_used),
		  notes       = COALESCE($22, notes),
		  updated_at  = now()
		 WHERE id = $1
		 RETURNING `+selectFields,
		id,
		req.Name, req.ClassName, req.Description, req.ColdBlood, req.Exhaustion,
		req.Abilities, req.Strength, req.Agility, req.Toughness, req.Presence, req.HP, req.MaxHP,
		req.Weapons, req.ArmorName, req.ArmorTier, req.Sufferings, req.Equipment,
		req.Silver, req.Omens, req.OmensUsed, req.Notes,
	)

	c, err := scanCharacter(row)
	if err != nil {
		return nil, fmt.Errorf("patch morkborg character: %w", err)
	}
	return c, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM morkborg_characters WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete morkborg character: %w", err)
	}
	return nil
}
