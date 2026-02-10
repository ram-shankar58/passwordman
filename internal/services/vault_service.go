package services

import (
	"context"
	"errors"
	"time"

	"vault/internal/models"
	"vault/internal/repository"
)

type VaultService struct {
	repo   *repository.VaultRepository
	crypto *CryptoService
	audit  *AuditService
}

func NewVaultService(repo *repository.VaultRepository, crypto *CryptoService, audit *AuditService) *VaultService {
	return &VaultService{repo: repo, crypto: crypto, audit: audit}
}

func (s *VaultService) List(userID int64) ([]models.VaultEntry, error) {
	entries, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		entries[i].Password = ""
	}
	return entries, nil
}

func (s *VaultService) Get(userID, id int64) (*models.VaultEntry, error) {
	entry, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, err
	}

	plain, err := s.crypto.Decrypt(entry.PasswordEnc)
	if err != nil {
		return nil, err
	}
	entry.Password = plain

	// Log access in background using goroutine (non-blocking)
	s.audit.LogEvent(userID, id, "accessed")

	// Update last accessed timestamp asynchronously
	go func() {
		_ = s.repo.TouchLastAccessed(userID, id, time.Now().UTC())
	}()

	return entry, nil
}

// Search finds vault entries by website/URL/username with context support
func (s *VaultService) Search(ctx context.Context, userID int64, query string) ([]models.VaultEntry, error) {
	// Use context for potential cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if query == "" {
		return s.List(userID)
	}

	entries, err := s.repo.SearchByWebsite(userID, query)
	if err != nil {
		return nil, err
	}

	// Clear passwords from search results
	for i := range entries {
		entries[i].Password = ""
	}

	return entries, nil
}

func (s *VaultService) Create(userID int64, entry models.VaultEntry) (int64, error) {
	if entry.Title == "" || entry.Password == "" {
		return 0, errors.New("title and password required")
	}

	enc, err := s.crypto.Encrypt(entry.Password)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	entry.UserID = userID
	entry.PasswordEnc = enc
	entry.CreatedAt = now
	entry.UpdatedAt = now

	return s.repo.Create(entry)
}

func (s *VaultService) Update(userID, id int64, entry models.VaultEntry) error {
	if entry.Title == "" {
		return errors.New("title required")
	}

	current, err := s.repo.GetByID(userID, id)
	if err != nil {
		return err
	}

	current.Title = entry.Title
	current.Username = entry.Username
	current.URL = entry.URL
	current.Category = entry.Category
	current.Notes = entry.Notes
	current.UpdatedAt = time.Now().UTC()

	if entry.Password != "" {
		enc, err := s.crypto.Encrypt(entry.Password)
		if err != nil {
			return err
		}
		current.PasswordEnc = enc
	}

	return s.repo.Update(*current)
}

func (s *VaultService) Delete(userID, id int64) error {
	return s.repo.Delete(userID, id)
}
