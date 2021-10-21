package integration

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/diagnostic/log"
	"github.com/pghq/go-museum/museum/store/postgres"
	"github.com/pghq/go-museum/museum/store/repository"
)

const (
	// DefaultContainerTTL is the default ttl for docker containers
	DefaultContainerTTL = time.Minute

	// DefaultMaxConnectTime is the default amount of time to allow connecting
	DefaultMaxConnectTime = 30 * time.Second

	// DefaultTag is the default tag for the postgres docker image
	DefaultTag = "11"

	// DefaultDockerEndpoint is the default docker endpoint for connections
	DefaultDockerEndpoint = ""
)

// Postgres is an integration for running postgres tests using docker
type Postgres struct {
	Repository *repository.Repository
	Migration  struct {
		FS        fs.FS
		Directory string
	}
	ImageTag     string
	ContainerTTL time.Duration
	MaxConnectTime time.Duration
	DockerEndpoint string

	exit func(code int)
	run func() int
	purge func(r *dockertest.Resource) error
	emit func(err error)
}

// NewPostgres creates a new integration test for postgres
func NewPostgres(m *testing.M) *Postgres{
	p := Postgres{
		run: m.Run,
		exit: os.Exit,
		emit: errors.Send,
	}

	return &p
}

// NewPostgresWithExit creates a new postgres image with an expected exit
func NewPostgresWithExit(t *testing.T, code int) *Postgres{
	p := Postgres{
		run: NoopRun,
		emit: errors.Send,
		exit: ExpectExit(t, code),
	}

	return &p
}

// RunPostgres runs a new postgres integration
func RunPostgres(integration *Postgres) {
	if integration.ContainerTTL == 0 {
		integration.ContainerTTL = DefaultContainerTTL
	}

	if integration.ImageTag == "" {
		integration.ImageTag = DefaultTag
	}

	if integration.MaxConnectTime == 0 {
		integration.MaxConnectTime = DefaultMaxConnectTime
	}

	if integration.DockerEndpoint == ""{
		integration.DockerEndpoint = DefaultDockerEndpoint
	}

	pool, err := dockertest.NewPool(integration.DockerEndpoint)
	if err != nil {
		integration.emit(err)
		integration.exit(1)
		return
	}

	pool.MaxWait = integration.MaxConnectTime
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
		integration.emit(err)
		integration.exit(1)
		return
	}

	// Unfortunately, this method does not do any error handling :(
	_ = resource.Expire(uint(integration.ContainerTTL.Seconds()))
	connect := func() error {
		log.Writer(io.Discard)
		defer log.Reset()

		primary := fmt.Sprintf("postgres://test:test@localhost:%s/test?sslmode=disable", resource.GetPort("5432/tcp"))
		store := postgres.NewStore(primary).Migrations(integration.Migration.FS, integration.Migration.Directory)
		if err = store.Connect(); err != nil {
			err = errors.Wrap(err)
			return err
		}

		integration.Repository, _ = repository.New(store)
		return nil
	}

	purge := pool.Purge
	if integration.purge != nil{
		purge = func(r *dockertest.Resource) error {
			_ = pool.Purge(r)
			return integration.purge(resource)
		}
	}

	if deadline := pool.Retry(connect); deadline != nil {
		integration.emit(err)
		_ = purge(resource)
		integration.exit(1)
		return
	}

	code := integration.run()

	if err := purge(resource); err != nil {
		integration.emit(err)
		integration.exit(1)
		return
	}

	integration.exit(code)
}
