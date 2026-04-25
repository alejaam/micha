package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	periodapproval "micha/backend/internal/domain/period_approval"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

type PeriodApprovalRepository struct {
	db *pgxpool.Pool
}

func NewPeriodApprovalRepository(db *pgxpool.Pool) PeriodApprovalRepository {
	return PeriodApprovalRepository{db: db}
}

func (r PeriodApprovalRepository) Save(ctx context.Context, a periodapproval.PeriodApproval) error {
	attrs := a.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO period_approvals (id, member_id, period_id, status, comment, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (member_id, period_id) DO UPDATE
			SET status = EXCLUDED.status, comment = EXCLUDED.comment, updated_at = EXCLUDED.updated_at`,
		string(attrs.ID), attrs.MemberID, attrs.PeriodID, string(attrs.Status), attrs.Comment, attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("period approval repository save: %w", err)
	}
	return nil
}

func (r PeriodApprovalRepository) GetByMemberAndPeriod(ctx context.Context, memberID, periodID string) (periodapproval.PeriodApproval, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, member_id, period_id, status, comment, created_at, updated_at
			FROM period_approvals
			WHERE member_id = $1 AND period_id = $2`,
		memberID, periodID,
	)
	a, err := scanPeriodApproval(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return periodapproval.PeriodApproval{}, shared.ErrNotFound
	}
	if err != nil {
		return periodapproval.PeriodApproval{}, fmt.Errorf("period approval repository getByMemberAndPeriod: %w", err)
	}
	return a, nil
}

func (r PeriodApprovalRepository) ListByPeriod(ctx context.Context, periodID string) ([]periodapproval.PeriodApproval, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, member_id, period_id, status, comment, created_at, updated_at
			FROM period_approvals
			WHERE period_id = $1`,
		periodID,
	)
	if err != nil {
		return nil, fmt.Errorf("period approval repository listByPeriod: %w", err)
	}
	defer rows.Close()

	approvals := make([]periodapproval.PeriodApproval, 0)
	for rows.Next() {
		a, scanErr := scanPeriodApproval(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("period approval repository listByPeriod: scan: %w", scanErr)
		}
		approvals = append(approvals, a)
	}
	return approvals, nil
}

func (r PeriodApprovalRepository) DeleteAllByPeriod(ctx context.Context, periodID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM period_approvals WHERE period_id = $1`, periodID)
	if err != nil {
		return fmt.Errorf("period approval repository deleteAllByPeriod: %w", err)
	}
	return nil
}

func scanPeriodApproval(r pgx.Row) (periodapproval.PeriodApproval, error) {
	var (
		id        string
		memberID  string
		periodID  string
		status    string
		comment   string
		createdAt time.Time
		updatedAt time.Time
	)

	if err := r.Scan(&id, &memberID, &periodID, &status, &comment, &createdAt, &updatedAt); err != nil {
		return periodapproval.PeriodApproval{}, err
	}

	return periodapproval.NewFromAttributes(periodapproval.PeriodApprovalAttributes{
		ID:        periodapproval.ID(id),
		MemberID:  memberID,
		PeriodID:  periodID,
		Status:    periodapproval.ApprovalStatus(status),
		Comment:   comment,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})
}

var _ outbound.PeriodApprovalRepository = PeriodApprovalRepository{}
