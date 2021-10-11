package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/internal/database"
)

// Query creates a query for the database.
func (db *Database) Query() database.Query{
	return NewQuery(db)
}

// Query is an instance of the repository query for Postgres.
type Query struct{
	db   *Database
	pool Pool
	opts []func(builder squirrel.SelectBuilder) squirrel.SelectBuilder
}

func (q *Query) Secondary() database.Query{
	if q.db != nil{
		q.pool = q.db.secondary
	}

	return q
}

func (q *Query) From(collection string) database.Query{
	q.opts = append(q.opts, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder{
		return builder.From(collection)
	})

	return q
}

func (q *Query) And(collection string, args ...interface{}) database.Query{
	q.opts = append(q.opts, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder{
		return builder.Join(collection, args...)
	})

	return q
}

func (q *Query) Filter(filter string, args ...interface{}) database.Query{
	q.opts = append(q.opts, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder{
		return builder.Where(filter, args...)
	})

	return q
}

func (q *Query) Order(by string) database.Query{
	q.opts = append(q.opts, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder{
		return builder.OrderBy(by)
	})

	return q
}

func (q *Query) First(first uint) database.Query{
	q.opts = append(q.opts, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder{
		return builder.Limit(uint64(first))
	})

	return q
}

func (q *Query) After(key string, value interface{}) database.Query{
	q.opts = append(q.opts, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder{
		return builder.Where(squirrel.GtOrEq{key: value})
	})

	return q
}

func (q *Query) Return(key string, args ...interface{}) database.Query{
	q.opts = append(q.opts, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder{
		return builder.Column(key, args...)
	})

	return q
}

func (q *Query) Statement() (string, []interface{}, error){
	builder := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Select()

	for _, opt := range q.opts{
		builder = opt(builder)
	}

	return builder.ToSql()
}

func (q *Query) Execute(ctx context.Context) (database.Cursor, error){
	sql, args, err := q.Statement()
	if err != nil{
		return nil, errors.BadRequest(err)
	}

	rows, err := q.pool.Query(ctx, sql, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NoContent(err)
		}

		return nil, errors.Wrap(err)
	}

	return NewCursor(rows), nil
}

func (q *Query) Decode(to database.Query) error{
	nq, ok := to.(*Query)
	if !ok{
		return errors.NewBadRequest("not a postgres query")
	}

	nq.db = q.db
	nq.pool = q.pool
	nq.opts = append(q.opts, nq.opts...)

	return nil
}

// NewQuery creates a new query for the Postgres database.
func NewQuery(db *Database) *Query {
	q := Query{
		db: db,
		pool: db.pool,
	}

	return &q
}
