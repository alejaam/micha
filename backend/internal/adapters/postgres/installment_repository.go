package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/installment"
	"micha/backend/internal/ports/outbound"
)

// InstallmentRepository fulfils outbound.InstallmentRepository using PostgreSQL.
type InstallmentRepository struct {
	db *pgxpool.Pool
}

// NewInstallmentRepository constructs an InstallmentRepository backed by the given pool.
func NewInstallmentRepository(db *pgxpool.Pool) InstallmentRepository {
	return InstallmentRepository{db: db}
}

// Save persists a single installment record.
func (r InstallmentRepository) Save(ctx context.Context, inst installment.Installment) error {
	attrs := inst.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO installments (id, expense_id, paid_by_member_id, start_date, installment_amount_cents, total_amount_cents, total_installments, current_installment, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		string(attrs.ID), attrs.ExpenseID, attrs.PaidByMemberID, attrs.StartDate,
		attrs.InstallmentAmountCents, attrs.TotalAmountCents, attrs.TotalInstallments,
		attrs.CurrentInstallment, attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("installment repository save: %w", err)
	}
	return nil
}

// SaveAll persists multiple installments in a single transaction.
func (r InstallmentRepository) SaveAll(ctx context.Context, insts []installment.Installment) error {
	if len(insts) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("installment repository saveAll begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, inst := range insts {
		attrs := inst.Attributes()
		_, err := tx.Exec(ctx,
			`INSERT INTO installments (id, expense_id, paid_by_member_id, start_date, installment_amount_cents, total_amount_cents, total_installments, current_installment, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			string(attrs.ID), attrs.ExpenseID, attrs.PaidByMemberID, attrs.StartDate,
			attrs.InstallmentAmountCents, attrs.TotalAmountCents, attrs.TotalInstallments,
			attrs.CurrentInstallment, attrs.CreatedAt, attrs.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("installment repository saveAll insert: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("installment repository saveAll commit: %w", err)
	}
	return nil
}

// ListByExpense returns all installments for a specific MSI expense.
func (r InstallmentRepository) ListByExpense(ctx context.Context, expenseID string) ([]installment.Installment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, expense_id, paid_by_member_id, start_date, installment_amount_cents, total_amount_cents, total_installments, current_installment, created_at, updated_at
			FROM installments
			WHERE expense_id = $1
			ORDER BY current_installment ASC`,
		expenseID,
	)
	if err != nil {
		return nil, fmt.Errorf("installment repository listByExpense: %w", err)
	}
	defer rows.Close()

	var installments []installment.Installment
	for rows.Next() {
		inst, err := scanInstallment(rows)
		if err != nil {
			return nil, fmt.Errorf("installment repository listByExpense scan: %w", err)
		}
		installments = append(installments, inst)
	}
	return installments, nil
}

// ListByHouseholdAndPeriod returns installments for a household in a given date range.
func (r InstallmentRepository) ListByHouseholdAndPeriod(ctx context.Context, householdID string, from, to time.Time) ([]installment.Installment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT i.id, i.expense_id, i.paid_by_member_id, i.start_date, i.installment_amount_cents, i.total_amount_cents, i.total_installments, i.current_installment, i.created_at, i.updated_at
			FROM installments i
			JOIN members m ON i.paid_by_member_id = m.id
			WHERE m.household_id = $1
				AND i.start_date >= $2
				AND i.start_date < $3
			ORDER BY i.start_date ASC`,
		householdID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("installment repository listByHouseholdAndPeriod: %w", err)
	}
	defer rows.Close()

	var installments []installment.Installment
	for rows.Next() {
		inst, err := scanInstallment(rows)
		if err != nil {
			return nil, fmt.Errorf("installment repository listByHouseholdAndPeriod scan: %w", err)
		}
		installments = append(installments, inst)
	}
	return installments, nil
}

// DeleteByExpense removes all installments for a given root expense.
func (r InstallmentRepository) DeleteByExpense(ctx context.Context, expenseID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM installments WHERE expense_id = $1`, expenseID)
	if err != nil {
		return fmt.Errorf("installment repository deleteByExpense: %w", err)
	}
	return nil
}

func scanInstallment(r pgx.Row) (installment.Installment, error) {
	var (
		id                     string
		expenseID              string
		paidByMemberID         string
		startDate              time.Time
		installmentAmountCents int64
		totalAmountCents       int64
		totalInstallments      int
		currentInstallment     int
		createdAt              time.Time
		updatedAt              time.Time
	)

	err := r.Scan(&id, &expenseID, &paidByMemberID, &startDate, &installmentAmountCents, &totalAmountCents, &totalInstallments, &currentInstallment, &createdAt, &updatedAt)
	if err != nil {
		return installment.Installment{}, err
	}

	return installment.NewFromAttributes(installment.InstallmentAttributes{
		ID:                     installment.ID(id),
		ExpenseID:              expenseID,
		PaidByMemberID:         paidByMemberID,
		StartDate:              startDate,
		InstallmentAmountCents: installmentAmountCents,
		TotalAmountCents:       totalAmountCents,
		TotalInstallments:      totalInstallments,
		CurrentInstallment:     currentInstallment,
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
	})
}

var _ outbound.InstallmentRepository = InstallmentRepository{}
