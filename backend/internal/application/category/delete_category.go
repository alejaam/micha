package categoryapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ErrCannotDeleteDefault is returned when attempting to delete a built-in category.
var ErrCannotDeleteDefault = errors.New("cannot delete a default category")

// DeleteCategoryUseCase deletes a custom (non-default) category from a household.
type DeleteCategoryUseCase struct {
	repo outbound.CategoryRepository
}

// NewDeleteCategoryUseCase constructs a DeleteCategoryUseCase.
func NewDeleteCategoryUseCase(repo outbound.CategoryRepository) DeleteCategoryUseCase {
	return DeleteCategoryUseCase{repo: repo}
}

func (u DeleteCategoryUseCase) Execute(ctx context.Context, input inbound.DeleteCategoryInput) error {
	cats, err := u.repo.ListByHousehold(ctx, input.HouseholdID)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	var found bool
	for _, c := range cats {
		if string(c.ID()) == input.CategoryID {
			found = true
			if c.IsDefault() {
				return fmt.Errorf("delete category: %w", ErrCannotDeleteDefault)
			}
			break
		}
	}
	if !found {
		return fmt.Errorf("delete category: %w", shared.ErrNotFound)
	}

	if err := u.repo.Delete(ctx, input.CategoryID); err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	slog.InfoContext(ctx, "delete category", "category_id", input.CategoryID, "household_id", input.HouseholdID)
	return nil
}

var _ inbound.DeleteCategoryUseCase = DeleteCategoryUseCase{}
