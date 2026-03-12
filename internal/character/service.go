package character

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID string) ([]ListItem, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, id int, userID string) (*Character, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("character not found")
		}
		return nil, err
	}
	if c.UserID != userID {
		return nil, fmt.Errorf("forbidden")
	}
	return c, nil
}

func (s *Service) Create(ctx context.Context, userID, name, clan string, data json.RawMessage) (*Character, error) {
	return s.repo.Create(ctx, userID, name, clan, data)
}

func (s *Service) Update(ctx context.Context, id int, userID, name, clan string, data json.RawMessage) (*Character, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("character not found")
		}
		return nil, err
	}
	if c.UserID != userID {
		return nil, fmt.Errorf("forbidden")
	}
	return s.repo.Update(ctx, id, name, clan, data)
}

func (s *Service) Delete(ctx context.Context, id int, userID string) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return fmt.Errorf("character not found")
		}
		return err
	}
	if c.UserID != userID {
		return fmt.Errorf("forbidden")
	}
	return s.repo.Delete(ctx, id)
}

func isNotFound(err error) bool {
	return err != nil && (err == pgx.ErrNoRows || containsNoRows(err))
}

func containsNoRows(err error) bool {
	return err != nil && len(err.Error()) > 0 &&
		(err.Error() == "no rows in result set" ||
			unwrapContains(err, pgx.ErrNoRows))
}

func unwrapContains(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		if u, ok := err.(unwrapper); ok {
			err = u.Unwrap()
		} else {
			return false
		}
	}
	return false
}
