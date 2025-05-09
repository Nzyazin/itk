package postgresdb

import (
	"context"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"github.com/Nzyazin/itk/internal/core/logger"
	"github.com/Nzyazin/itk/pkg/config"
)

type Database struct {
	log logger.Logger
	*sqlx.DB
}

func NewPostgresDB(cfg config.DBConfig, log logger.Logger) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
	)

	db, err := sqlx.Open("postgres", connStr)
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

	return &Database{log: log, DB: db}, nil
}

func (db *Database) Close() error {
	db.log.Info("Closing database connection")
	return db.DB.Close()
}
