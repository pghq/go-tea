package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"

	"github.com/pghq/go-museum/museum/errors"
	"github.com/pghq/go-museum/museum/internal/database"
)

func (db *Database) Remove() database.Remove{
	return NewRemove(db)
}

type Remove struct {
	db *Database
	opts []func(builder squirrel.DeleteBuilder) squirrel.DeleteBuilder
}

func (r *Remove) From(collection string) database.Remove{
	r.opts = append(r.opts, func(builder squirrel.DeleteBuilder) squirrel.DeleteBuilder{
		return builder.From(collection)
	})

	return r
}

func (r *Remove) Filter(filter string, args ...interface{}) database.Remove{
	r.opts = append(r.opts, func(builder squirrel.DeleteBuilder) squirrel.DeleteBuilder{
		return builder.Where(filter, args...)
	})

	return r
}

func (r *Remove) Order(by string) database.Remove{
	r.opts = append(r.opts, func(builder squirrel.DeleteBuilder) squirrel.DeleteBuilder{
		return builder.OrderBy(by)
	})

	return r
}

func (r *Remove) First(first uint) database.Remove{
	r.opts = append(r.opts, func(builder squirrel.DeleteBuilder) squirrel.DeleteBuilder{
		return builder.Limit(uint64(first))
	})

	return r
}

func (r *Remove) After(key string, value interface{}) database.Remove{
	r.opts = append(r.opts, func(builder squirrel.DeleteBuilder) squirrel.DeleteBuilder{
		return builder.Where(squirrel.GtOrEq{key: value})
	})

	return r
}

func (r *Remove) Decode(to database.Remove) error{
	nr, ok := to.(*Remove)
	if !ok{
		return errors.NewBadRequest("not a postgres remove command")
	}

	nr.db = r.db
	nr.opts = append(r.opts, nr.opts...)

	return nil
}

func (r *Remove) Execute(ctx context.Context) (uint, error){
	sql, args, err := r.Statement()
	if err != nil{
		return 0, errors.BadRequest(err)
	}

	tag, err := r.db.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err)
	}

	return uint(tag.RowsAffected()), nil
}

func (r *Remove) Statement() (string, []interface{}, error){
	builder := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Delete("")

	for _, opt := range r.opts{
		builder = opt(builder)
	}

	return builder.ToSql()
}

func NewRemove(db *Database) *Remove{
	r := Remove{db: db}
	return &r
}
