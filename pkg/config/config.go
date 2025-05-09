package config

import (
	"os"
	"strconv"
	"fmt"
	
	"path/filepath"
	"github.com/joho/godotenv"
)


type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	MaxOpenConns int
    MaxIdleConns int
}

func LoadConfigDB() (*DBConfig, error) {
	err := godotenv.Load(filepath.Join("config.env"))
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
    if err != nil {
        return nil, fmt.Errorf("invalid DB_PORT: %w", err)
    }

    maxOpen, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
    if err != nil {
        return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
    }

    maxIdle, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
    if err != nil {
        return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
    }

	return &DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		MaxOpenConns: maxOpen,
		MaxIdleConns: maxIdle,
	}, nil
}