package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"vault/internal/config"
)

func Open(cfg config.Config) (*sql.DB, error) {
	dir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("file:%s?_pgrama=busy_timeout(5000)", cfg.DBPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil

}

func Migrate(db *sql.DB) error {
	migration, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(migration))
	return err
}
