package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Nzyazin/zadnik.store/internal/auth/config"
	_ "github.com/lib/pq"
	"time"
)

type Config struct {
    Host         string
    Port         int
    User         string
    Password     string
    DBName       string
    SSLMode      string
    MaxOpenConns int
    MaxIdleConns int
}

type Database struct {
	*sql.DB
}

func NewPostgresDB(cfg Config, log logger.Logger) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(2 * time.Hour)

	return &Database{db}, nil
}