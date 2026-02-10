package repository

import (
	"database/sql"
	"time"

	"vault/internal/models"
)

type VaultRepository struct {
	db *sql.DB
}

func NewVaultRepository(db *sql.DB) *VaultRepository {
	return &VaultRepository{db: db}
}

func (r *VaultRepository) ListByUser(userID int64) ([]models.VaultEntry, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, title, username, password_enc, url, category, notes, created_at, updated_at, last_accessed_at FROM vault_entries WHERE user_id = ? ORDER BY id DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.VaultEntry{}
	for rows.Next() {
		entry, err := scanVaultEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, rows.Err()
}

func (r *VaultRepository) GetByID(userID, id int64) (*models.VaultEntry, error) {
	row := r.db.QueryRow(
		"SELECT id, user_id, title, username, password_enc, url, category, notes, created_at, updated_at, last_accessed_at FROM vault_entries WHERE user_id = ? AND id = ?",
		userID,
		id,
	)
	return scanVaultEntry(row)
}

func (r *VaultRepository) Create(entry models.VaultEntry) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO vault_entries (user_id, title, username, password_enc, url, category, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.UserID,
		entry.Title,
		entry.Username,
		entry.PasswordEnc,
		entry.URL,
		entry.Category,
		entry.Notes,
		entry.CreatedAt.UTC().Format(time.RFC3339),
		entry.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *VaultRepository) Update(entry models.VaultEntry) error {
	_, err := r.db.Exec(
		`UPDATE vault_entries
		SET title = ?, username = ?, password_enc = ?, url = ?, category = ?, notes = ?, updated_at = ?
		WHERE user_id = ? AND id = ?`,
		entry.Title,
		entry.Username,
		entry.PasswordEnc,
		entry.URL,
		entry.Category,
		entry.Notes,
		entry.UpdatedAt.UTC().Format(time.RFC3339),
		entry.UserID,
		entry.ID,
	)
	return err
}

func (r *VaultRepository) Delete(userID, id int64) error {
	_, err := r.db.Exec("DELETE FROM vault_entries WHERE user_id = ? AND id = ?", userID, id)
	return err
}

func (r *VaultRepository) SearchByWebsite(userID int64, query string) ([]models.VaultEntry, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, title, username, password_enc, url, category, notes, created_at, updated_at, last_accessed_at 
		FROM vault_entries 
		WHERE user_id = ? AND (title LIKE ? OR url LIKE ? OR username LIKE ?)
		ORDER BY id DESC`,
		userID,
		"%"+query+"%",
		"%"+query+"%",
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.VaultEntry{}
	for rows.Next() {
		entry, err := scanVaultEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, rows.Err()
}

func (r *VaultRepository) TouchLastAccessed(userID, id int64, accessedAt time.Time) error {
	_, err := r.db.Exec(
		"UPDATE vault_entries SET last_accessed_at = ? WHERE user_id = ? AND id = ?",
		accessedAt.UTC().Format(time.RFC3339),
		userID,
		id,
	)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanVaultEntry(row scanner) (*models.VaultEntry, error) {
	var entry models.VaultEntry
	var createdAt string
	var updatedAt string
	var lastAccessed sql.NullString

	err := row.Scan(
		&entry.ID,
		&entry.UserID,
		&entry.Title,
		&entry.Username,
		&entry.PasswordEnc,
		&entry.URL,
		&entry.Category,
		&entry.Notes,
		&createdAt,
		&updatedAt,
		&lastAccessed,
	)
	if err != nil {
		return nil, err
	}

	entry.CreatedAt = parseTime(createdAt)
	entry.UpdatedAt = parseTime(updatedAt)
	if lastAccessed.Valid {
		t := parseTime(lastAccessed.String)
		entry.LastAccessedAt = &t
	}

	return &entry, nil
}
