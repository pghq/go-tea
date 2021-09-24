package postgres

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/errors"
	"github.com/pghq/go-museum/museum/internal/test"
	"github.com/pghq/go-museum/museum/log"
)

func TestNew(t *testing.T) {
	t.Run("NotNil", func(t *testing.T) {
		dsn := "postgres://postgres:postgres@db:5432"
		db := New(dsn)
		assert.Equal(t, dsn, db.primaryDSN)
		assert.NotNil(t, dsn, db.secondaryDSN)
		assert.Equal(t, DefaultSQLMaxOpenConns, db.maxConns)
		assert.Equal(t, time.Duration(0), db.maxConnLifetime)
	})
}

func TestDatabase_MaxConnLifetime(t *testing.T) {
	db := New("postgres://postgres:postgres@db:5432")

	t.Run("NotNil", func(t *testing.T) {
		lifetime := time.Second
		db = db.MaxConnLifetime(lifetime)
		assert.NotNil(t, db)
		assert.Equal(t, lifetime, db.maxConnLifetime)
	})
}

func TestDatabase_MaxConns(t *testing.T) {
	db := New("postgres://postgres:postgres@db:5432")

	t.Run("NotNil", func(t *testing.T) {
		conns := 5
		db = db.MaxConns(conns)
		assert.NotNil(t, db)
		assert.Equal(t, conns, db.maxConns)
	})
}

func TestDatabase_Connect(t *testing.T) {
	t.Run("NotNil", func(t *testing.T) {
		db := New("")
		err := db.Connect()
		assert.NotNil(t, err)
		assert.False(t, db.IsConnected())
	})

	t.Run("BadSecondary", func(t *testing.T) {
		db := New("postgres://postgres:postgres@db:5432")
		db.connect = func(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
			if config.ConnString() == "secondary"{
				return nil, errors.New("bad secondary")
			}
			return &pgxpool.Pool{}, nil
		}
		err := db.Secondary("secondary").Connect()
		assert.NotNil(t, err)
		assert.False(t, db.IsConnected())
	})

	t.Run("NoError", func(t *testing.T) {
		db, _, _ := setup(t)
		assert.NotNil(t, db.pool)
		assert.NotNil(t, db.secondary)
		assert.True(t, db.IsConnected())
	})
}

func TestDatabase_Add(t *testing.T) {
	db, primary, _ := setup(t)
	ctx := context.TODO()

	t.Run("NotNil", func(t *testing.T) {
		add := db.Add()
		assert.NotNil(t, add)
	})

	t.Run("BadRequestError", func(t *testing.T) {
		_, err := db.Add().Execute(context.TODO())
		assert.NotNil(t, err)
	})

	t.Run("Decode", func(t *testing.T) {
		from := NewAdd(db)
		from.To("tests")
		to := NewAdd(&Database{})
		err := from.Decode(to)
		assert.Nil(t, err)
		assert.Equal(t, from.opts, to.opts)
		assert.Equal(t, from.db, to.db)
	})

	t.Run("DecodeError", func(t *testing.T) {
		add := db.Add()
		err := add.Decode(nil)
		assert.NotNil(t, err)
		assert.False(t, errors.IsFatal(err))
	})

	t.Run("FatalError", func(t *testing.T) {
		primary.Error(errors.New("an error has occurred"))
		_, err := db.Add().
			To("tests").
			Item(map[string]interface{}{"coverage": 50}).
			Execute(context.TODO())
		assert.NotNil(t, err)
		assert.True(t, errors.IsFatal(err))
	})

	t.Run("IntegrityError", func(t *testing.T) {
		primary.Error(&pgconn.PgError{Code: pgerrcode.IntegrityConstraintViolation})
		_, err := db.Add().
			To("tests").
			Item(map[string]interface{}{"coverage": 50}).
			Execute(context.TODO())
		assert.NotNil(t, err)
		assert.False(t, errors.IsFatal(err))
	})

	t.Run("NoError", func(t *testing.T) {
		primary.ExpectExec(ctx, "INSERT INTO tests (coverage) SELECT coverage FROM units LIMIT 1")
		defer primary.Assert()
		_, err := db.Add().
			Item(map[string]interface{}{"coverage": 0}).
			Query(db.Query().From("units").Return("coverage").First(1)).
			To("tests").
			Execute(context.TODO())
		assert.Nil(t, err)
	})
}

func TestDatabase_Query(t *testing.T) {
	db, primary, secondary := setup(t)
	ctx := context.TODO()

	t.Run("NotNil", func(t *testing.T) {
		query := db.Query()
		assert.NotNil(t, query)
	})

	t.Run("BadRequestError", func(t *testing.T) {
		_, err := db.Query().Execute(context.TODO())
		assert.NotNil(t, err)
	})

	t.Run("Decode", func(t *testing.T) {
		from := NewQuery(db)
		from.From("tests")
		to := NewQuery(&Database{})
		err := from.Decode(to)
		assert.Nil(t, err)
		assert.Equal(t, from.opts, to.opts)
		assert.Equal(t, from.db, to.db)
		assert.Equal(t, from.pool, to.pool)
	})

	t.Run("DecodeError", func(t *testing.T) {
		query := db.Query()
		err := query.Decode(nil)
		assert.NotNil(t, err)
		assert.False(t, errors.IsFatal(err))
	})

	t.Run("NoContentError", func(t *testing.T) {
		primary.Error(pgx.ErrNoRows)
		_, err := db.Query().
			From("tests").
			Return("runs").
			Execute(context.TODO())
		assert.NotNil(t, err)
		assert.False(t, errors.IsFatal(err))
	})

	t.Run("FatalError", func(t *testing.T) {
		primary.Error(errors.New("an error has occurred"))
		_, err := db.Query().
			From("tests").
			Return("runs").
			Execute(context.TODO())
		assert.NotNil(t, err)
		assert.True(t, errors.IsFatal(err))
	})

	t.Run("Primary", func(t *testing.T) {
		primary.ExpectQuery(ctx, "SELECT runs FROM tests JOIN units ON runs.id = units.id WHERE coverage > $1 AND id >= $2 ORDER BY coverage DESC LIMIT 5", 50, 2)
		defer primary.Assert()
		_, err := db.Query().
			From("tests").
			And("units ON runs.id = units.id").
			Filter("coverage > ?", 50).
			Order("coverage DESC").
			Return("runs").
			First(5).
			After("id", 2).
			Execute(context.TODO())
		assert.Nil(t, err)
	})

	t.Run("Secondary", func(t *testing.T) {
		secondary.ExpectQuery(ctx, "SELECT runs FROM tests")
		defer secondary.Assert()
		_, err := db.Query().
			Secondary().
			From("tests").
			Return("runs").
			Execute(context.TODO())
		assert.Nil(t, err)
	})
}

func TestDatabase_Remove(t *testing.T) {
	db, primary, _ := setup(t)
	ctx := context.TODO()

	t.Run("NotNil", func(t *testing.T) {
		remove := db.Remove()
		assert.NotNil(t, remove)
	})

	t.Run("BadRequestError", func(t *testing.T) {
		_, err := db.Remove().Execute(context.TODO())
		assert.NotNil(t, err)
	})

	t.Run("Decode", func(t *testing.T) {
		from := NewRemove(db)
		from.From("tests")
		to := NewRemove(&Database{})
		err := from.Decode(to)
		assert.Nil(t, err)
		assert.Equal(t, from.opts, to.opts)
		assert.Equal(t, from.db, to.db)
	})

	t.Run("DecodeError", func(t *testing.T) {
		remove := db.Remove()
		err := remove.Decode(nil)
		assert.NotNil(t, err)
		assert.False(t, errors.IsFatal(err))
	})

	t.Run("FatalError", func(t *testing.T) {
		primary.Error(errors.New("an error has occurred"))
		_, err := db.Remove().
			From("tests").
			Execute(context.TODO())
		assert.NotNil(t, err)
		assert.True(t, errors.IsFatal(err))
	})

	t.Run("NoError", func(t *testing.T) {
		primary.ExpectExec(ctx, "DELETE FROM tests WHERE coverage > $1 AND id >= $2 ORDER BY coverage DESC LIMIT 5", 50, 2)
		defer primary.Assert()
		_, err := db.Remove().
			From("tests").
			Filter("coverage > ?", 50).
			Order("coverage DESC").
			First(5).
			After("id", 2).
			Execute(context.TODO())
		assert.Nil(t, err)
	})
}

func TestDatabase_Transaction(t *testing.T) {
	db, primary, _ := setup(t)
	ctx := context.TODO()

	t.Run("BeginError", func(t *testing.T) {
		primary.Error(errors.New("an error occurred"))
		_, err := db.Transaction(ctx)
		assert.NotNil(t, err)
	})
	
	t.Run("NoError", func(t *testing.T) {
		primary.ExpectBegin(ctx)
		defer primary.Assert()
		tx, err := db.Transaction(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, tx)
	})
}

func TestTransaction_Commit(t *testing.T) {
	db, primary, _ := setup(t)
	ctx := context.TODO()

	t.Run("AnError", func(t *testing.T) {
		ptx := test.NewPostgresTx(t).
			Error(errors.New("an error has occurred"))
		defer ptx.Assert()
		primary.ExpectBegin(ctx).ReturnBegin(ptx)
		defer primary.Assert()
		tx, _ := db.Transaction(ctx)
		err := tx.Commit()
		assert.NotNil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		ptx := test.NewPostgresTx(t).ExpectCommit(ctx)
		defer ptx.Assert()
		primary.ExpectBegin(ctx).ReturnBegin(ptx)
		defer primary.Assert()
		tx, _ := db.Transaction(ctx)
		err := tx.Commit()
		assert.Nil(t, err)
	})
}

func TestTransaction_Execute(t *testing.T) {
	db, primary, _ := setup(t)
	ctx := context.TODO()

	t.Run("BadRequestError", func(t *testing.T) {
		ptx := test.NewPostgresTx(t)
		defer ptx.Assert()
		primary.ExpectBegin(ctx).ReturnBegin(ptx)
		defer primary.Assert()
		tx, _ := db.Transaction(ctx)
		_, err := tx.Execute(db.Remove())
		assert.NotNil(t, err)
	})

	t.Run("FatalError", func(t *testing.T) {
		ptx := test.NewPostgresTx(t).Error(errors.New("an error has occurred"))
		defer ptx.Assert()
		primary.ExpectBegin(ctx).ReturnBegin(ptx)
		defer primary.Assert()
		tx, _ := db.Transaction(ctx)
		_, err := tx.Execute(db.Remove().From("tests"))
		assert.NotNil(t, err)
		assert.True(t, errors.IsFatal(err))
	})

	t.Run("Primary", func(t *testing.T) {
		ptx := test.NewPostgresTx(t).ExpectExec(ctx, "DELETE FROM tests WHERE coverage > $1 AND id >= $2 ORDER BY coverage DESC LIMIT 5", 50, 2)
		defer ptx.Assert()
		primary.ExpectBegin(ctx).ReturnBegin(ptx)
		defer primary.Assert()
		tx, _ := db.Transaction(ctx)
		_, err := tx.Execute(db.Remove().
			From("tests").
			Filter("coverage > ?", 50).
			Order("coverage DESC").
			First(5).
			After("id", 2),
		)
		assert.Nil(t, err)
	})
}

func TestTransaction_Rollback(t *testing.T) {
	db, primary, _ := setup(t)
	ctx := context.TODO()

	t.Run("AnError", func(t *testing.T) {
		ptx := test.NewPostgresTx(t).Error(errors.New("an error has occurred"))
		defer ptx.Assert()
		primary.ExpectBegin(ctx).ReturnBegin(ptx)
		defer primary.Assert()
		tx, _ := db.Transaction(ctx)
		err := tx.Rollback()
		assert.NotNil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		ptx := test.NewPostgresTx(t).ExpectRollback(ctx)
		defer ptx.Assert()
		primary.ExpectBegin(ctx).ReturnBegin(ptx)
		defer primary.Assert()
		tx, _ := db.Transaction(ctx)
		err := tx.Rollback()
		assert.Nil(t, err)
	})
}

func TestNewCursor(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		rows := test.NewPostgresRows(t)
		defer rows.Assert()
		c := NewCursor(rows)
		assert.NotNil(t, c)
	})
}

func TestCursor_Close(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		rows := test.NewPostgresRows(t).ExpectClose()
		defer rows.Assert()
		c := NewCursor(rows)
		defer c.Close()
	})
}

func TestCursor_Decode(t *testing.T) {
	t.Run("DecodeError", func(t *testing.T) {
		rows := test.NewPostgresRows(t).
			Error(errors.New("an error has occurred"))
		defer rows.Assert()
		c := NewCursor(rows)
		assert.NotNil(t, c)
		err := c.Decode(nil)
		assert.NotNil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		var one int
		rows := test.NewPostgresRows(t).ExpectScan(&one)
		defer rows.Assert()
		c := NewCursor(rows)
		err := c.Decode(&one)
		assert.Nil(t, err)
		assert.Nil(t, c.Error())
	})
}

func TestCursor_Error(t *testing.T) {
	t.Run("AnError", func(t *testing.T) {
		rows := test.NewPostgresRows(t).
			Error(errors.New("an error has occurred"))
		defer rows.Assert()
		c := NewCursor(rows)
		err := c.Error()
		assert.NotNil(t, err)
	})
}

func TestCursor_Next(t *testing.T) {
	t.Run("MakesCall", func(t *testing.T) {
		rows := test.NewPostgresRows(t).
			ExpectNext()
		defer rows.Assert()
		c := NewCursor(rows)
		c.Next()
	})
}

func TestIsIntegrityConstraintViolation(t *testing.T) {
	t.Run("Yes", func(t *testing.T) {
		err := &pgconn.PgError{Code: pgerrcode.IntegrityConstraintViolation}
		assert.True(t, IsIntegrityConstraintViolation(err))
	})

	t.Run("No", func(t *testing.T) {
		err := errors.New("an error has occurred")
		assert.False(t, IsIntegrityConstraintViolation(err))
	})
}

func TestLogger_Log(t *testing.T) {
	l := NewLogger()
	t.Run("Debug", func(t *testing.T){
		log.Level("debug")
		var buf bytes.Buffer
		log.Writer(&buf)
		l.Log(context.TODO(), pgx.LogLevelDebug, "an error has occurred", nil)
		assert.True(t, strings.Contains(buf.String(), "debug"))
	})

	t.Run("Info", func(t *testing.T){
		log.Level("info")
		var buf bytes.Buffer
		log.Writer(&buf)
		l.Log(context.TODO(), pgx.LogLevelInfo, "an error has occurred", nil)
		assert.True(t, strings.Contains(buf.String(), "info"))
	})

	t.Run("Warn", func(t *testing.T){
		log.Level("warn")
		var buf bytes.Buffer
		log.Writer(&buf)
		l.Log(context.TODO(), pgx.LogLevelWarn, "an error has occurred", nil)
		assert.True(t, strings.Contains(buf.String(), "warn"))
	})

	t.Run("Error", func(t *testing.T){
		log.Level("error")
		var buf bytes.Buffer
		log.Writer(&buf)
		l.Log(context.TODO(), pgx.LogLevelError, "an error has occurred", nil)
		assert.True(t, strings.Contains(buf.String(), "error"))
	})
}

func setup(t *testing.T) (*Database, *test.PostgresPool, *test.PostgresPool){
	t.Helper()

	db := New("postgres://postgres:postgres@db:5432")
	db.connect = func(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
		t.Helper()
		assert.NotNil(t, config)
		assert.IsType(t, &Logger{}, config.ConnConfig.Logger)
		assert.Equal(t, int32(DefaultSQLMaxOpenConns), config.MaxConns)
		assert.Equal(t, time.Duration(0), config.MaxConnLifetime)
		assert.Equal(t, db.primaryDSN, config.ConnString())

		return &pgxpool.Pool{}, nil
	}
	err := db.Connect()
	assert.Nil(t, err)
	primary := test.NewPostgresPool(t)
	secondary := test.NewPostgresPool(t)
	db.pool = primary
	db.secondary = secondary

	return db, primary, secondary
}