package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/pghq/go-museum/museum/errors"
	"github.com/pghq/go-museum/museum/internal/database"
	"github.com/pghq/go-museum/museum/log"
)

const (
	// DefaultSQLMaxOpenConns is the default maximum number of open connections.
	DefaultSQLMaxOpenConns = 100
)

type Database struct {
	pool Pool
	secondary Pool
	primaryDSN string
	secondaryDSN string
	maxConns int
	maxConnLifetime time.Duration
	connect func(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error)
}

// MaxConns sets the max number of open connections.
func (db *Database) MaxConns(conns int) *Database {
	db.maxConns = conns

	return db
}

// MaxConnLifetime sets the max lifetime for a connection.
func (db *Database) MaxConnLifetime(timeout time.Duration) *Database {
	db.maxConnLifetime = timeout

	return db
}

func (db *Database) Connect() error{
	primary, err := db.newPool(db.primaryDSN)
	if err != nil {
		return errors.Wrap(err)
	}

	db.pool = primary
	secondary, err := db.newPool(db.secondaryDSN)
	if err != nil {
		return errors.Wrap(err)
	}

	db.secondary = secondary

	return nil
}

func (db *Database) IsConnected() bool{
	return db.pool != nil && db.secondary != nil
}

// newPool creates a new concurrency safe pool
func (db *Database) newPool(databaseURL string) (Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	config.ConnConfig.Logger = NewLogger()
	config.MaxConnLifetime = db.maxConnLifetime
	config.MaxConns = int32(db.maxConns)

	pool, err := db.connect(context.Background(), config)
	if err != nil{
		return nil, errors.Wrap(err)
	}

	return pool, nil
}

func (db *Database) Secondary(dsn string) *Database{
	if dsn != ""{
		db.secondaryDSN = dsn
	}

	return db
}

func New(primary string) *Database {
	db := Database{
		primaryDSN: primary,
		secondaryDSN: primary,
		connect: pgxpool.ConnectConfig,
	}
	db.maxConns = DefaultSQLMaxOpenConns

	return &db
}

// Pool for executing db commands against.
type Pool interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Cursor represents an instance of a Cursor
type Cursor struct{
	dest []interface{}
	rows pgx.Rows
}

func (c *Cursor) Next() bool{
	return c.rows.Next()
}

func (c *Cursor) Decode(values ...interface{}) error {
	if err := c.rows.Scan(values...); err != nil{
		return errors.Wrap(err)
	}

	return nil
}

func (c *Cursor) Close(){
	c.rows.Close()
}

func (c *Cursor) Error() error{
	if err := c.rows.Err(); err != nil{
		return errors.Wrap(err)
	}

	return nil
}

// NewCursor constructs a new cursor instance.
func NewCursor(rows pgx.Rows) database.Cursor {
	return &Cursor{
		rows: rows,
	}
}

// Logger is an instance of the pgx Logger
type Logger struct {}

func (l *Logger) Log(_ context.Context, level pgx.LogLevel, msg string, _ map[string]interface{}){
	switch level{
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
func NewLogger() *Logger{
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
