package restorer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	backuper "github.com/oiler-backup/postgres-adapter/backuper/backuper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	ctx        = context.Background()
	dbUser     = "testuser"
	dbPass     = "testpass"
	dbName     = "testdb"
	backupName = "backup.dump"
)

func setupPostgresContainer() (*tc.Container, error) {
	req := tc.ContainerRequest{
		Image:           "postgres:14",
		ExposedPorts:    []string{"5432/tcp"},
		AlwaysPullImage: false,
		Env: map[string]string{
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPass,
			"POSTGRES_DB":       dbName,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	mysqlC, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	return &mysqlC, err
}

func Test_Redtore_UploadValidDump(t *testing.T) {
	postgresC, err := setupPostgresContainer()
	require.NoError(t, err)
	defer func() {
		err := (*postgresC).Terminate(ctx)
		if err != nil {
			panic(err)
		}
	}()

	dbhost, _ := (*postgresC).ContainerIP(ctx)
	dbPort, _ := (*postgresC).MappedPort(ctx, "5432")
	tempDir := t.TempDir()
	backupFile := filepath.Join(tempDir, "backup.dump")

	b := backuper.NewBackuper(
		dbhost,
		dbPort.Port(),
		dbUser,
		dbPass,
		dbName,
		backupFile,
	)

	err = b.Backup(ctx, false)

	r := NewRestorer(
		dbhost,
		dbPort.Port(),
		dbUser,
		dbPass,
		dbName,
		backupFile,
	)

	err = r.Restore(ctx)
	require.NoError(t, err)

	fileInfo, err := os.Stat(backupFile)
	require.NoError(t, err)
	assert.Greater(t, fileInfo.Size(), int64(0))
}

func Test_Redtore_InvalidDump(t *testing.T) {
	postgresC, err := setupPostgresContainer()
	require.NoError(t, err)
	defer func() {
		err := (*postgresC).Terminate(ctx)
		if err != nil {
			panic(err)
		}
	}()

	dbhost, _ := (*postgresC).ContainerIP(ctx)
	dbPort, _ := (*postgresC).MappedPort(ctx, "5432")
	tempDir := t.TempDir()
	backupFile := filepath.Join(tempDir, backupName)

	r := NewRestorer(
		dbhost,
		dbPort.Port(),
		dbUser,
		dbPass,
		dbName,
		backupFile,
	)

	err = r.Restore(ctx)
	require.ErrorContains(t, err, "failed executing mysql restore:")
}

func Test_Redtore_InvalidDBHost(t *testing.T) {
	dbhost := "wrong"
	dbPort := "5432"
	r := NewRestorer(
		dbhost,
		dbPort,
		dbUser,
		dbPass,
		dbName,
		backupName,
	)

	err := r.Restore(ctx)
	require.ErrorContains(t, err, "failed to connect to database:")
}
