package integration

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/store/postgres"
	"github.com/pghq/go-museum/museum/store/repository"
)

const (
	// DefaultContainerTTL is the default ttl for docker containers
	DefaultContainerTTL = time.Minute

	// DefaultTag is the default tag for the postgres docker image
	DefaultTag = "11"
)

// Postgres is an integration for running postgres tests using docker
type Postgres struct {
	Main       *testing.M
	Repository *repository.Repository
	Migration  struct {
		fs        fs.FS
		directory string
	}
	ImageTag     string
	ContainerTTL time.Duration
}

// RunPostgres runs a new postgres integration
func RunPostgres(integration *Postgres) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		errors.Emit(err)
		os.Exit(1)
	}

	if integration.ContainerTTL == 0 {
		integration.ContainerTTL = DefaultContainerTTL
	}

	if integration.ImageTag == "" {
		integration.ImageTag = DefaultTag
	}

	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        integration.ImageTag,
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_DB=test",
			"listen_addresses='*'",
		},
	}

	resource, err := pool.RunWithOptions(&opts, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	if err != nil {
		errors.Emit(err)
		os.Exit(1)
	}

	if err := resource.Expire(uint(integration.ContainerTTL.Seconds())); err != nil {
		errors.Emit(err)
		os.Exit(1)
	}

	connect := func() error {
		primary := fmt.Sprintf("postgres://test:test@localhost:%s/test?sslmode=disable", resource.GetPort("5432/tcp"))
		store := postgres.NewStore(primary).Migrations(integration.Migration.fs, integration.Migration.directory)
		if err := store.Connect(); err != nil {
			return errors.Wrap(err)
		}

		var err error
		integration.Repository, err = repository.New(store)
		if err != nil {
			return err
		}

		return nil
	}

	if err := pool.Retry(connect); err != nil {
		errors.Emit(err)
		os.Exit(1)
	}

	code := integration.Main.Run()

	if err := pool.Purge(resource); err != nil {
		errors.Emit(err)
		os.Exit(1)
	}

	os.Exit(code)
}
