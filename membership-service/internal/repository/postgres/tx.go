package postgres

import (
	"context"
	"errors"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type contextKey string

const txKey contextKey = "tx"

type txManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) domain.TxManager {
	return &txManager{db: db}
}

func (m *txManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txKey, tx)

	err = fn(txCtx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func getQueryer(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx, ok := ctx.Value(txKey).(*sqlx.Tx); ok {
		return tx
	}
	return db
}
