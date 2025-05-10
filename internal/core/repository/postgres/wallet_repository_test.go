package postgres_test

import (
	"context"
	"testing"
	"time"
	"fmt"
	"sync"

	"github.com/Nzyazin/itk/internal/core/models"
	"github.com/Nzyazin/itk/internal/core/repository/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/google/uuid"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types"
	"github.com/Nzyazin/itk/internal/core/logger"
    "github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T, log logger.Logger) (*sqlx.DB, func()) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.41"))
	if err != nil {
		log.Error("Failed to create Docker client", logger.ErrorField("error", err))
		t.Fatalf("Failed to create Docker client: %v", err)
	}

	ctx := context.Background()
	containerName := "postgres_test_db"

	port := "5433"
	portBindings := nat.PortMap{
		"5433/tcp": []nat.PortBinding{{HostPort: port}},
	}

	containerConfig := &container.Config{
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_DB=test_db",
		},
	}
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
	}
    _ = cli.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{Force: true})

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		log.Error("Failed to create container", logger.ErrorField("error", err))
		t.Fatalf("Failed to create container: %v", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Error("Failed to start container", logger.ErrorField("error", err))
		t.Fatalf("Failed to start container: %v", err)
	}

	stopContainer := func() {
		if err := cli.ContainerStop(ctx, resp.ID, container.StopOptions{}); err != nil {
			log.Error("Failed to stop container", logger.ErrorField("error", err))
			t.Fatalf("Failed to stop container: %v", err)
		}
		if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
			log.Error("Failed to remove container", logger.ErrorField("error", err))
			t.Fatalf("Failed to remove container: %v", err)
		}
	}

	db, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://test:test@localhost:%s/test_db?sslmode=disable", port))
	if err != nil {
		log.Error("Failed to connect to PostgreSQL", logger.ErrorField("error", err))
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Error("Failed to ping PostgreSQL", logger.ErrorField("error", err))
		t.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	return db, stopContainer
}

func TestConcurrentDeposits(t *testing.T) {
	log, cleanup := logger.NewLogger()
	defer cleanup()

	db, teardown := setupTestDB(t, log)
	defer teardown()

	repo := postgres.NewPostgresWalletRepo(db, log) // передаем логгер в репозиторий

	// Создаем кошелек
	walletID := uuid.New()
	_, err := db.Exec(`INSERT INTO wallets (id, balance, currency_code, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())`, walletID, 0, "USD")
	require.NoError(t, err)

	// Параметры нагрузки
	const goroutines = 1000
	const amount = int64(1)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errCh := make(chan error, goroutines)

	ctx := context.Background()

	start := time.Now()

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			_, err := repo.ExecuteTxWithRetry(ctx, walletID, amount, models.OperationDeposit)
			if err != nil {
				log.Error(fmt.Sprintf("transaction %d failed", i), logger.ErrorField("error", err))
			}
			errCh <- err
		}(i)
	}

	wg.Wait()
	close(errCh)

	var errorCount int
	for err := range errCh {
		if err != nil {
			log.Error("transaction failed", logger.ErrorField("error", err))
			errorCount++
		}
	}

	var balance int64
	err = db.Get(&balance, "SELECT balance FROM wallets WHERE id = $1", walletID)
	require.NoError(t, err)

	assert.Equal(t, int64(goroutines), balance)
	assert.Equal(t, 0, errorCount, "some requests failed")

	t.Logf("Completed in %s", time.Since(start))
}
