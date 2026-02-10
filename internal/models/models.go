package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type VaultEntry struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"userId"`
	Title          string     `json:"title"`
	Username       string     `json:"username"`
	Password       string     `json:"password,omitempty"`
	PasswordEnc    string     `json:"-"`
	URL            string     `json:"url,omitempty"`
	Category       string     `json:"category,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	LastAccessedAt *time.Time `json:"lastAccessedAt,omitempty"`
}
