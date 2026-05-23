package subscriptioncatalogapp

import (
	"context"
	"fmt"

	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// GetSubscriptionKPIUseCase implements inbound.GetSubscriptionKPIUseCase.
type GetSubscriptionKPIUseCase struct {
	repo outbound.SubscriptionCatalogRepository
}

// NewGetSubscriptionKPIUseCase constructs a new GetSubscriptionKPIUseCase.
func NewGetSubscriptionKPIUseCase(repo outbound.SubscriptionCatalogRepository) GetSubscriptionKPIUseCase {
	return GetSubscriptionKPIUseCase{repo: repo}
}

// Execute calculates subscription KPI for a given household.
func (u GetSubscriptionKPIUseCase) Execute(ctx context.Context, householdID string) (inbound.SubscriptionKPIOutput, error) {
	links, err := u.repo.ListLinksByHousehold(ctx, householdID)
	if err != nil {
		return inbound.SubscriptionKPIOutput{}, fmt.Errorf("get subscription KPI: %w", err)
	}

	if len(links) == 0 {
		return inbound.SubscriptionKPIOutput{}, nil
	}

	// Calculate totals
	var totalSpentCents int64
	var standaloneTotalCents int64

	// Track unique expense IDs to avoid double-counting total_spent
	expenseAmounts := make(map[string]int64)      // expenseID → amountCents
	serviceCountByExpense := make(map[string]int) // expenseID → count of linked services

	// For overlap detection: serviceID → set of expenseIDs
	overlapMap := make(map[string]map[string]string) // serviceID → expenseID → serviceName

	for _, link := range links {
		price := link.StandalonePriceCents
		if link.CustomPriceCents != nil && *link.CustomPriceCents > 0 {
			price = *link.CustomPriceCents
		}
		standaloneTotalCents += price

		expenseAmounts[link.RecurringExpenseID] = link.ExpenseAmountCents
		serviceCountByExpense[link.RecurringExpenseID]++

		if _, ok := overlapMap[link.ServiceID]; !ok {
			overlapMap[link.ServiceID] = make(map[string]string)
		}
		overlapMap[link.ServiceID][link.RecurringExpenseID] = link.ServiceName
	}

	// Total spent: sum of expense amounts. Each expense is counted once.
	for _, amount := range expenseAmounts {
		totalSpentCents += amount
	}

	// Net savings
	netSavingsCents := standaloneTotalCents - totalSpentCents

	// Build overlap warnings
	var overlapWarnings []inbound.OverlapWarning
	for serviceID, expenseIDs := range overlapMap {
		if len(expenseIDs) > 1 {
			var serviceName string
			var ids []string
			for eID, sName := range expenseIDs {
				serviceName = sName
				ids = append(ids, eID)
			}
			overlapWarnings = append(overlapWarnings, inbound.OverlapWarning{
				CatalogServiceID:    serviceID,
				ServiceName:         serviceName,
				RecurringExpenseIDs: ids,
			})
		}
	}

	return inbound.SubscriptionKPIOutput{
		TotalSpentCents:      totalSpentCents,
		StandaloneTotalCents: standaloneTotalCents,
		NetSavingsCents:      netSavingsCents,
		OverlapWarnings:      overlapWarnings,
	}, nil
}

var _ inbound.GetSubscriptionKPIUseCase = GetSubscriptionKPIUseCase{}
