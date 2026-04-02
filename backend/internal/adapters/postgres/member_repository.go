// Package postgres provides PostgreSQL adapter implementations for outbound ports.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	var userID *string
	if attrs.UserID != "" {
		userID = &attrs.UserID
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO members (id, household_id, name, email, monthly_salary_cents, user_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		string(attrs.ID), attrs.HouseholdID, attrs.Name, attrs.Email, attrs.MonthlySalaryCents, userID, attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return shared.ErrAlreadyExists
		}
		return fmt.Errorf("member repository save: %w", err)
	}

	return nil
}

// FindByID retrieves a member by ID.
func (r MemberRepository) FindByID(ctx context.Context, id string) (member.Member, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, household_id, name, email, monthly_salary_cents, user_id, created_at, updated_at
			FROM members
			WHERE id = $1 AND deleted_at IS NULL`,
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

// FindByUserID returns the member linked to the given user within a household.
func (r MemberRepository) FindByUserID(ctx context.Context, householdID, userID string) (member.Member, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, household_id, name, email, monthly_salary_cents, user_id, created_at, updated_at
			FROM members
			WHERE household_id = $1 AND user_id = $2 AND deleted_at IS NULL
			LIMIT 1`,
		householdID, userID,
	)

	m, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return member.Member{}, shared.ErrNotFound
	}
	if err != nil {
		return member.Member{}, fmt.Errorf("member repository findByUserID: %w", err)
	}

	return m, nil
}

// ListByHousehold returns members for a household ordered by created_at DESC.
func (r MemberRepository) ListByHousehold(ctx context.Context, householdID string, limit, offset int) ([]member.Member, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, name, email, monthly_salary_cents, user_id, created_at, updated_at
			FROM members
			WHERE household_id = $1 AND deleted_at IS NULL
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

// ListAllByHousehold returns all members for a household ordered by created_at ASC.
func (r MemberRepository) ListAllByHousehold(ctx context.Context, householdID string) ([]member.Member, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, name, email, monthly_salary_cents, user_id, created_at, updated_at
			FROM members
			WHERE household_id = $1 AND deleted_at IS NULL
			ORDER BY created_at ASC`,
		householdID,
	)
	if err != nil {
		return nil, fmt.Errorf("member repository listAllByHousehold: %w", err)
	}
	defer rows.Close()

	members := make([]member.Member, 0)
	for rows.Next() {
		m, scanErr := scanMember(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("member repository listAllByHousehold: scan: %w", scanErr)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("member repository listAllByHousehold: rows: %w", err)
	}

	return members, nil
}

// Update persists changes to an existing member.
func (r MemberRepository) Update(ctx context.Context, m member.Member) error {
	attrs := m.Attributes()
	var userID *string
	if attrs.UserID != "" {
		userID = &attrs.UserID
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE members
			SET name = $1,
				email = $2,
				monthly_salary_cents = $3,
				user_id = $4,
				updated_at = $5
			WHERE id = $6`,
		attrs.Name, attrs.Email, attrs.MonthlySalaryCents, userID, attrs.UpdatedAt, string(attrs.ID),
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

// FindByUserIDGlobal returns any member record linked to the given user, regardless of household.
func (r MemberRepository) FindByUserIDGlobal(ctx context.Context, userID string) (member.Member, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, household_id, name, email, monthly_salary_cents, user_id, created_at, updated_at
			FROM members
			WHERE user_id = $1 AND deleted_at IS NULL
			LIMIT 1`,
		userID,
	)

	m, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return member.Member{}, shared.ErrNotFound
	}
	if err != nil {
		return member.Member{}, fmt.Errorf("member repository findByUserIDGlobal: %w", err)
	}

	return m, nil
}

// ListHouseholdIDsByUserID returns all household IDs the user belongs to.
func (r MemberRepository) ListHouseholdIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT household_id FROM members WHERE user_id = $1 AND deleted_at IS NULL`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("member repository listHouseholdIDsByUserID: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("member repository listHouseholdIDsByUserID: scan: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("member repository listHouseholdIDsByUserID: rows: %w", err)
	}

	return ids, nil
}

// Delete soft-deletes a member by setting deleted_at.
func (r MemberRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE members SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("member repository delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// CountActiveByHousehold returns the count of non-deleted members in a household.
func (r MemberRepository) CountActiveByHousehold(ctx context.Context, householdID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM members WHERE household_id = $1 AND deleted_at IS NULL`,
		householdID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("member repository countActiveByHousehold: %w", err)
	}
	return count, nil
}

func scanMember(r row) (member.Member, error) {
	var (
		id                 string
		householdID        string
		name               string
		email              string
		monthlySalaryCents int64
		userID             *string
		createdAt          time.Time
		updatedAt          time.Time
	)

	if err := r.Scan(&id, &householdID, &name, &email, &monthlySalaryCents, &userID, &createdAt, &updatedAt); err != nil {
		return member.Member{}, err
	}

	resolvedUserID := ""
	if userID != nil {
		resolvedUserID = *userID
	}

	return member.NewFromAttributes(member.Attributes{
		ID:                 member.ID(id),
		HouseholdID:        householdID,
		Name:               name,
		Email:              email,
		MonthlySalaryCents: monthlySalaryCents,
		UserID:             resolvedUserID,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	})
}
