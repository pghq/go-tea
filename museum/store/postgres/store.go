package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/diagnostic/log"
	"github.com/pghq/go-museum/museum/store"
)

const (
	// DefaultSQLMaxOpenConns is the default maximum number of open connections.
	DefaultSQLMaxOpenConns = 100
)

// Store is a client for interacting with Postgres.
type Store struct {
	pool       Pool
	secondary  Pool
	primaryDSN string
	secondaryDSN    string
	maxConns        int
	maxConnLifetime time.Duration
	connect         func(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error)
}

// MaxConns sets the max number of open connections.
func (s *Store) MaxConns(conns int) *Store {
	s.maxConns = conns

	return s
}

// MaxConnLifetime sets the max lifetime for a connection.
func (s *Store) MaxConnLifetime(timeout time.Duration) *Store {
	s.maxConnLifetime = timeout

	return s
}

func (s *Store) Connect() error {
	primary, err := s.newPool(s.primaryDSN)
	if err != nil {
		return errors.Wrap(err)
	}

	s.pool = primary
	secondary, err := s.newPool(s.secondaryDSN)
	if err != nil {
		return errors.Wrap(err)
	}

	s.secondary = secondary

	return nil
}

func (s *Store) IsConnected() bool {
	return s.pool != nil && s.secondary != nil
}

// newPool creates a new concurrency safe pool
func (s *Store) newPool(databaseURL string) (Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	config.ConnConfig.Logger = NewLogger()
	config.MaxConnLifetime = s.maxConnLifetime
	config.MaxConns = int32(s.maxConns)

	pool, err := s.connect(context.Background(), config)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return pool, nil
}

func (s *Store) Secondary(dsn string) *Store {
	if dsn != "" {
		s.secondaryDSN = dsn
	}

	return s
}

// NewStore creates a new Postgres database client.
func NewStore(primary string) *Store {
	s := Store{
		primaryDSN:   primary,
		secondaryDSN: primary,
		connect:      pgxpool.ConnectConfig,
	}
	s.maxConns = DefaultSQLMaxOpenConns

	return &s
}

// Pool for executing db commands against.
type Pool interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Cursor represents an instance of a Cursor
type Cursor struct {
	dest []interface{}
	rows pgx.Rows
}

func (c *Cursor) Next() bool {
	return c.rows.Next()
}

func (c *Cursor) Decode(values ...interface{}) error {
	if err := c.rows.Scan(values...); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (c *Cursor) Close() {
	c.rows.Close()
}

func (c *Cursor) Error() error {
	if err := c.rows.Err(); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// NewCursor constructs a new cursor instance.
func NewCursor(rows pgx.Rows) store.Cursor {
	return &Cursor{
		rows: rows,
	}
}

// Logger is an instance of the pgx Logger
type Logger struct{}

func (l *Logger) Log(_ context.Context, level pgx.LogLevel, msg string, _ map[string]interface{}) {
	switch level {
	case pgx.LogLevelDebug:
		log.Debug(msg)
	case pgx.LogLevelInfo:
		log.Info(msg)
	case pgx.LogLevelWarn:
		log.Warn(msg)
	default:
		log.Error(errors.New(msg))
	}
}

// NewLogger creates a new database pgx Logger
func NewLogger() *Logger {
	return &Logger{}
}

// IsIntegrityConstraintViolation checks if the error is an integrity constraint violation.
func IsIntegrityConstraintViolation(err error) bool {
	var e *pgconn.PgError
	if errors.As(err, &e) && pgerrcode.IsIntegrityConstraintViolation(e.Code) {
		return true
	}

	return false
}
