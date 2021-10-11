package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/internal/database"
)

// Add creates an add command for the database.
func (db *Database) Add() database.Add{
	return NewAdd(db)
}

// Add is an instance of the add repository command using Postgres.
type Add struct {
	db *Database
	opts []func(builder squirrel.InsertBuilder) squirrel.InsertBuilder
}

func (a *Add) To(collection string) database.Add{
	a.opts = append(a.opts, func(builder squirrel.InsertBuilder) squirrel.InsertBuilder{
		return builder.Into(collection)
	})

	return a
}

func (a *Add) Item(value map[string]interface{}) database.Add{
	a.opts = append(a.opts, func(builder squirrel.InsertBuilder) squirrel.InsertBuilder{
		return builder.SetMap(value)
	})

	return a
}

func (a *Add) Query(q database.Query) database.Add{
	if q, ok := q.(*Query); ok{
		s := squirrel.StatementBuilder.
			PlaceholderFormat(squirrel.Dollar).
			Select()
		for _, opt := range q.opts{
			s = opt(s)
		}

		a.opts = append(a.opts, func(builder squirrel.InsertBuilder) squirrel.InsertBuilder{
			return builder.Select(s)
		})
	}

	return a
}

func (a *Add) Decode(to database.Add) error{
	na, ok := to.(*Add)
	if !ok{
		return errors.NewBadRequest("not a postgres add command")
	}

	na.db = a.db
	na.opts = append(a.opts, na.opts...)

	return nil
}

func (a *Add) Execute(ctx context.Context) (uint, error){
	sql, args, err := a.Statement()
	if err != nil{
		return 0, errors.BadRequest(err)
	}

	tag, err := a.db.pool.Exec(ctx, sql, args...)
	if err != nil {
		if IsIntegrityConstraintViolation(err) {
			return 0, errors.BadRequest(err)
		}
		return 0, errors.Wrap(err)
	}

	return uint(tag.RowsAffected()), nil
}

func (a *Add) Statement() (string, []interface{}, error){
	builder := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Insert("")

	for _, opt := range a.opts{
		builder = opt(builder)
	}

	return builder.ToSql()
}

// NewAdd creates a new add command for the Postgres database.
func NewAdd(db *Database) *Add {
	a := Add{db: db}
	return &a
}
