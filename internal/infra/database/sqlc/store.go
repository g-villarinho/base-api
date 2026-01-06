package sqlc

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	ExecTx(ctx context.Context, fn func(Querier) error) error
}

type StoreImpl struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &StoreImpl{
		db:      db,
		Queries: New(db),
	}
}

func (store *StoreImpl) ExecTx(ctx context.Context, fn func(Querier) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := store.WithTx(tx)

	err = fn(q)

	if err != nil {

		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
