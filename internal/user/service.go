package user

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	usernameRegex     = regexp.MustCompile(`^[a-z0-9_-]+$`)
	reservedUsernames = map[string]bool{
		"admin": true, "root": true, "cainite": true,
		"api": true, "auth": true, "static": true,
	}
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, email, password, username string) (*User, error) {
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	taken, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	if taken {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u, err := s.repo.Create(ctx, email, string(hash), &username)
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}
	return u, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	if u.PasswordHash == nil {
		return nil, fmt.Errorf("login: invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("login: invalid credentials")
	}

	return u, nil
}

func (s *Service) UpsertGoogleUser(ctx context.Context, googleID, email, displayName string, avatarURL *string) (*User, error) {
	// Try by Google ID first
	u, err := s.repo.GetByGoogleID(ctx, googleID)
	if err == nil {
		return u, nil
	}

	// Try by email
	u, err = s.repo.GetByEmail(ctx, email)
	if err == nil {
		if u.GoogleID == nil {
			if err := s.repo.SetGoogleID(ctx, u.ID, googleID); err != nil {
				return nil, err
			}
		}
		return u, nil
	}

	// Create new user
	username, err := s.generateUniqueUsername(ctx, displayName)
	if err != nil {
		return nil, fmt.Errorf("generate username: %w", err)
	}

	return s.repo.CreateOAuth(ctx, email, googleID, username, avatarURL)
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, username *string, avatarURL *string) (*User, error) {
	if username != nil {
		if err := ValidateUsername(*username); err != nil {
			return nil, err
		}
		taken, err := s.repo.UsernameExists(ctx, *username)
		if err != nil {
			return nil, fmt.Errorf("update profile: %w", err)
		}
		if taken {
			return nil, ErrUsernameTaken
		}
	}
	return s.repo.UpdateProfile(ctx, userID, username, avatarURL)
}

func (s *Service) generateUniqueUsername(ctx context.Context, displayName string) (*string, error) {
	base := sanitizeUsername(displayName)
	if len(base) < 3 {
		base = "user"
	}

	candidate := base
	for i := 2; i <= 100; i++ {
		exists, err := s.repo.UsernameExists(ctx, candidate)
		if err != nil {
			return nil, err
		}
		if !exists {
			return &candidate, nil
		}
		candidate = fmt.Sprintf("%s_%d", base, i)
	}
	return nil, fmt.Errorf("could not generate unique username")
}

func sanitizeUsername(name string) string {
	lower := strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9_-]`)
	clean := re.ReplaceAllString(lower, "_")
	clean = regexp.MustCompile(`_+`).ReplaceAllString(clean, "_")
	clean = strings.Trim(clean, "_-")
	if len(clean) > 30 {
		clean = clean[:30]
	}
	return clean
}

func ValidateUsername(username string) error {
	if len(username) < 3 {
		return ErrUsernameTooShort
	}
	if len(username) > 30 {
		return ErrUsernameTooLong
	}
	if !usernameRegex.MatchString(username) {
		return ErrUsernameInvalid
	}
	if reservedUsernames[username] {
		return ErrUsernameInvalid
	}
	return nil
}

var (
	ErrUsernameTaken    = fmt.Errorf("USERNAME_TAKEN")
	ErrUsernameInvalid  = fmt.Errorf("USERNAME_INVALID")
	ErrUsernameTooShort = fmt.Errorf("USERNAME_TOO_SHORT")
	ErrUsernameTooLong  = fmt.Errorf("USERNAME_TOO_LONG")
)
