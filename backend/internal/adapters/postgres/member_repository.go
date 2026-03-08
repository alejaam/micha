// Package postgres provides PostgreSQL adapter implementations for outbound ports.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

// MemberRepository fulfils outbound.MemberRepository using PostgreSQL.
type MemberRepository struct {
	db *pgxpool.Pool
}

// NewMemberRepository constructs a MemberRepository backed by the given pool.
func NewMemberRepository(db *pgxpool.Pool) MemberRepository {
	return MemberRepository{db: db}
}

// Save persists a new member record.
func (r MemberRepository) Save(ctx context.Context, m member.Member) error {
	attrs := m.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO members (id, household_id, name, email, monthly_salary_cents, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		string(attrs.ID), attrs.HouseholdID, attrs.Name, attrs.Email, attrs.MonthlySalaryCents, attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("member repository save: %w", err)
	}

	return nil
}

// FindByID retrieves a member by ID.
func (r MemberRepository) FindByID(ctx context.Context, id string) (member.Member, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, household_id, name, email, monthly_salary_cents, created_at, updated_at
			FROM members
			WHERE id = $1`,
		id,
	)

	m, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return member.Member{}, shared.ErrNotFound
	}
	if err != nil {
		return member.Member{}, fmt.Errorf("member repository findByID: %w", err)
	}

	return m, nil
}

// ListByHousehold returns members for a household ordered by created_at DESC.
func (r MemberRepository) ListByHousehold(ctx context.Context, householdID string, limit, offset int) ([]member.Member, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, name, email, monthly_salary_cents, created_at, updated_at
			FROM members
			WHERE household_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`,
		householdID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("member repository listByHousehold: %w", err)
	}
	defer rows.Close()

	var members []member.Member
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, fmt.Errorf("member repository listByHousehold: scan: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("member repository listByHousehold: rows: %w", err)
	}

	return members, nil
}

// Update persists changes to an existing member.
func (r MemberRepository) Update(ctx context.Context, m member.Member) error {
	attrs := m.Attributes()
	tag, err := r.db.Exec(ctx,
		`UPDATE members
			SET name = $1,
				email = $2,
				monthly_salary_cents = $3,
				updated_at = $4
			WHERE id = $5`,
		attrs.Name, attrs.Email, attrs.MonthlySalaryCents, attrs.UpdatedAt, string(attrs.ID),
	)
	if err != nil {
		return fmt.Errorf("member repository update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}

	return nil
}

// ensure interface compliance at compile time.
var _ outbound.MemberRepository = MemberRepository{}

func scanMember(r row) (member.Member, error) {
	var (
		id                 string
		householdID        string
		name               string
		email              string
		monthlySalaryCents int64
		createdAt          time.Time
		updatedAt          time.Time
	)

	if err := r.Scan(&id, &householdID, &name, &email, &monthlySalaryCents, &createdAt, &updatedAt); err != nil {
		return member.Member{}, err
	}

	return member.NewFromAttributes(member.Attributes{
		ID:                 member.ID(id),
		HouseholdID:        householdID,
		Name:               name,
		Email:              email,
		MonthlySalaryCents: monthlySalaryCents,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	})
}
