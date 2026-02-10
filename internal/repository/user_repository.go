package repository

import (
	"database/sql"
	"time"

	"vault/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(email, passwordHash string) (int64, error) {
	res, err := r.db.Exec(
		"INSERT INTO users (email, password_hash, created_at) VALUES (?, ?, ?)",
		email,
		passwordHash,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE email = ?",
		email,
	)

	var user models.User
	var createdAt string
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &createdAt); err != nil {
		return nil, err
	}
	user.CreatedAt = parseTime(createdAt)
	return &user, nil
}

func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	row := r.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE id = ?",
		id,
	)

	var user models.User
	var createdAt string
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &createdAt); err != nil {
		return nil, err
	}
	user.CreatedAt = parseTime(createdAt)
	return &user, nil
}

func parseTime(value string) time.Time {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Now().UTC()
}
