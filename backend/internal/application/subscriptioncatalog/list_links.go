package subscriptioncatalogapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/subscriptioncatalog"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ListLinksByExpenseUseCase implements inbound.ListLinksByExpenseUseCase.
type ListLinksByExpenseUseCase struct {
	repo outbound.SubscriptionCatalogRepository
}

// NewListLinksByExpenseUseCase constructs a new ListLinksByExpenseUseCase.
func NewListLinksByExpenseUseCase(repo outbound.SubscriptionCatalogRepository) ListLinksByExpenseUseCase {
	return ListLinksByExpenseUseCase{repo: repo}
}

// Execute returns all catalog links for a given recurring expense.
func (u ListLinksByExpenseUseCase) Execute(ctx context.Context, recurringExpenseID string) ([]subscriptioncatalog.CatalogLink, error) {
	links, err := u.repo.ListLinksByExpense(ctx, recurringExpenseID)
	if err != nil {
		return nil, fmt.Errorf("list links by expense: %w", err)
	}
	return links, nil
}

var _ inbound.ListLinksByExpenseUseCase = ListLinksByExpenseUseCase{}
