package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

type PeriodRepository struct {
	db *pgxpool.Pool
}

func NewPeriodRepository(db *pgxpool.Pool) PeriodRepository {
	return PeriodRepository{db: db}
}

func (r PeriodRepository) getQuerier(ctx context.Context) Querier {
	return querierFromContext(ctx, r.db)
}

func (r PeriodRepository) Create(ctx context.Context, p period.Period) error {
	attrs := p.Attributes()
	_, err := r.getQuerier(ctx).Exec(ctx,
		`INSERT INTO periods (id, household_id, start_date, end_date, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		string(attrs.ID), attrs.HouseholdID, attrs.StartDate, attrs.EndDate, string(attrs.Status), attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("period repository create: %w", err)
	}
	return nil
}

func (r PeriodRepository) Update(ctx context.Context, p period.Period) error {
	attrs := p.Attributes()
	tag, err := r.getQuerier(ctx).Exec(ctx,
		`UPDATE periods
			SET status = $1, updated_at = $2
			WHERE id = $3`,
		string(attrs.Status), attrs.UpdatedAt, string(attrs.ID),
	)
	if err != nil {
		return fmt.Errorf("period repository update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r PeriodRepository) GetByID(ctx context.Context, id period.ID) (period.Period, error) {
	row := r.getQuerier(ctx).QueryRow(ctx,
		`SELECT id, household_id, start_date, end_date, status, created_at, updated_at
			FROM periods
			WHERE id = $1`,
		string(id),
	)
	p, err := scanPeriod(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return period.Period{}, shared.ErrNotFound
	}
	if err != nil {
		return period.Period{}, fmt.Errorf("period repository getByID: %w", err)
	}
	return p, nil
}

func (r PeriodRepository) GetCurrentOpen(ctx context.Context, householdID string) (period.Period, error) {
	row := r.getQuerier(ctx).QueryRow(ctx,
		`SELECT id, household_id, start_date, end_date, status, created_at, updated_at
			FROM periods
			WHERE household_id = $1 AND status = 'open'
			ORDER BY created_at DESC
			LIMIT 1`,
		householdID,
	)
	p, err := scanPeriod(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return period.Period{}, shared.ErrNotFound
	}
	if err != nil {
		return period.Period{}, fmt.Errorf("period repository getCurrentOpen: %w", err)
	}
	return p, nil
}

func (r PeriodRepository) GetLatestByHousehold(ctx context.Context, householdID string) (period.Period, error) {
	row := r.getQuerier(ctx).QueryRow(ctx,
		`SELECT id, household_id, start_date, end_date, status, created_at, updated_at
			FROM periods
			WHERE household_id = $1
			ORDER BY start_date DESC
			LIMIT 1`,
		householdID,
	)
	p, err := scanPeriod(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return period.Period{}, shared.ErrNotFound
	}
	if err != nil {
		return period.Period{}, fmt.Errorf("period repository getLatestByHousehold: %w", err)
	}
	return p, nil
}

func (r PeriodRepository) ListByHousehold(ctx context.Context, householdID string, limit, offset int) ([]period.Period, error) {
	rows, err := r.getQuerier(ctx).Query(ctx,
		`SELECT id, household_id, start_date, end_date, status, created_at, updated_at
			FROM periods
			WHERE household_id = $1
			ORDER BY start_date DESC
			LIMIT $2 OFFSET $3`,
		householdID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("period repository listByHousehold: %w", err)
	}
	defer rows.Close()

	var periods []period.Period
	for rows.Next() {
		p, err := scanPeriod(rows)
		if err != nil {
			return nil, fmt.Errorf("period repository listByHousehold: scan: %w", err)
		}
		periods = append(periods, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("period repository listByHousehold: rows: %w", err)
	}
	return periods, nil
}

func scanPeriod(r pgx.Row) (period.Period, error) {
	var (
		id          string
		householdID string
		startDate   time.Time
		endDate     time.Time
		status      string
		createdAt   time.Time
		updatedAt   time.Time
	)

	if err := r.Scan(&id, &householdID, &startDate, &endDate, &status, &createdAt, &updatedAt); err != nil {
		return period.Period{}, err
	}

	return period.NewFromAttributes(period.PeriodAttributes{
		ID:          period.ID(id),
		HouseholdID: householdID,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      period.Status(status),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	})
}

var _ outbound.PeriodRepository = PeriodRepository{}
