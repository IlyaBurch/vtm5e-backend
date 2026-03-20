package user

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     *string   `json:"username"`
	AvatarURL    *string   `json:"avatarUrl"`
	GoogleID     *string   `json:"-"`
	PasswordHash *string   `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}
