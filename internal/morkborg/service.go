package morkborg

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID string) ([]*Character, error) {
	chars, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if chars == nil {
		chars = []*Character{}
	}
	return chars, nil
}

func (s *Service) Create(ctx context.Context, userID string, req *CreateRequest) (*Character, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return s.repo.Create(ctx, userID, req)
}

func (s *Service) Get(ctx context.Context, id, userID string) (*Character, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.UserID != userID {
		return nil, ErrForbidden
	}
	return c, nil
}

func (s *Service) Patch(ctx context.Context, id, userID string, req *PatchRequest) (*Character, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	if c.UserID != userID {
		return nil, ErrForbidden
	}
	return s.repo.Patch(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if c.UserID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, id)
}

var (
	ErrNotFound  = fmt.Errorf("NOT_FOUND")
	ErrForbidden = fmt.Errorf("FORBIDDEN")
)
