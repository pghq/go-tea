package postgres

import (
	"context"

	"github.com/jackc/pgx/v4"

	"github.com/pghq/go-museum/museum/errors"
	"github.com/pghq/go-museum/museum/internal/database"
)

func (db *Database) Transaction(ctx context.Context) (database.Transaction, error){
	tx, err := db.pool.Begin(ctx)
	if err != nil{
		return nil, errors.Wrap(err)
	}

	t := transaction{ctx: ctx, tx: tx}
	return &t, err
}

type transaction struct {
	ctx context.Context
	tx pgx.Tx
}

func (t *transaction) Execute(statement database.Communicator) (uint, error){
	sql, args, err := statement.Statement()
	if err != nil{
		return 0, errors.BadRequest(err)
	}

	tag, err := t.tx.Exec(t.ctx, sql, args...)
	if err != nil{
		return 0, errors.Wrap(err)
	}

	return uint(tag.RowsAffected()), nil
}

func (t *transaction) Commit() error{
	if err := t.tx.Commit(t.ctx); err != nil{
		return errors.Wrap(err)
	}

	return nil

}
func (t *transaction) Rollback() error{
	if err := t.tx.Rollback(t.ctx); err != nil{
		return errors.Wrap(err)
	}

	return nil
}
