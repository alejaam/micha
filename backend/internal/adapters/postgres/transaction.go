package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier defines the minimal interface for executing SQL queries.
// Both *pgxpool.Pool and pgx.Tx satisfy this interface.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// contextKey is a private type for storing transaction context values.
type contextKey struct{}

// TransactionManager implements outbound.TransactionManager using pgx.
type TransactionManager struct {
	db *pgxpool.Pool
}

// NewTransactionManager creates a new TransactionManager.
func NewTransactionManager(db *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{db: db}
}

// Run starts a transaction, stores it in the context, calls fn, then commits or rolls back.
func (tm *TransactionManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction begin: %w", err)
	}

	txCtx := context.WithValue(ctx, contextKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(err, fmt.Errorf("transaction rollback: %w", rbErr))
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit: %w", err)
	}

	return nil
}

// querierFromContext returns a pgx.Tx from the context if one exists, otherwise returns the pool.
func querierFromContext(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := ctx.Value(contextKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}
