package subscriptioncatalogapp

import (
	"context"
	"testing"

	"micha/backend/internal/ports/outbound"
)

// --- Mock Repository ---

type mockSubscriptionCatalogRepo struct {
	outbound.SubscriptionCatalogRepository
	links []outbound.RecurringExpenseLink
	err   error
}

func (m *mockSubscriptionCatalogRepo) ListLinksByHousehold(_ context.Context, _ string) ([]outbound.RecurringExpenseLink, error) {
	return m.links, m.err
}

func TestGetSubscriptionKPIUseCase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		links            []outbound.RecurringExpenseLink
		wantTotalSpent   int64
		wantStandalone   int64
		wantSavings      int64
		wantOverlapCount int
	}{
		{
			name: "single expense single service — no overlap, positive savings",
			links: []outbound.RecurringExpenseLink{
				{
					RecurringExpenseID:   "exp-1",
					ServiceID:            "svc-1",
					ServiceName:          "Netflix Premium",
					StandalonePriceCents: 29900,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   21900,
				},
			},
			wantTotalSpent:   21900,
			wantStandalone:   29900,
			wantSavings:      8000, // 29900 - 21900
			wantOverlapCount: 0,
		},
		{
			name: "single expense multiple services — sum standalone, positive savings",
			links: []outbound.RecurringExpenseLink{
				{
					RecurringExpenseID:   "exp-1",
					ServiceID:            "svc-1",
					ServiceName:          "Netflix Premium",
					StandalonePriceCents: 29900,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   32900,
				},
				{
					RecurringExpenseID:   "exp-1",
					ServiceID:            "svc-2",
					ServiceName:          "Spotify Familiar",
					StandalonePriceCents: 19900,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   32900,
				},
			},
			wantTotalSpent:   32900, // counted once per expense
			wantStandalone:   49800, // 29900 + 19900
			wantSavings:      16900, // 49800 - 32900
			wantOverlapCount: 0,
		},
		{
			name: "custom price override used instead of standalone",
			links: []outbound.RecurringExpenseLink{
				{
					RecurringExpenseID:   "exp-1",
					ServiceID:            "svc-1",
					ServiceName:          "Apple TV+",
					StandalonePriceCents: 6900,
					CustomPriceCents:     int64Ptr(7900),
					ExpenseAmountCents:   32900,
				},
			},
			wantTotalSpent:   32900,
			wantStandalone:   7900,   // custom price overrides standalone
			wantSavings:      -25000, // 7900 - 32900 = negative savings
			wantOverlapCount: 0,
		},
		{
			name: "negative savings when bundle costs more than standalone",
			links: []outbound.RecurringExpenseLink{
				{
					RecurringExpenseID:   "exp-1",
					ServiceID:            "svc-1",
					ServiceName:          "Expensive Bundle",
					StandalonePriceCents: 10000,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   15000,
				},
			},
			wantTotalSpent:   15000,
			wantStandalone:   10000,
			wantSavings:      -5000,
			wantOverlapCount: 0,
		},
		{
			name: "overlap detection — same service linked to two expenses",
			links: []outbound.RecurringExpenseLink{
				{
					RecurringExpenseID:   "exp-1",
					ServiceID:            "svc-1",
					ServiceName:          "Apple TV+",
					StandalonePriceCents: 6900,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   32900,
				},
				{
					RecurringExpenseID:   "exp-2",
					ServiceID:            "svc-1",
					ServiceName:          "Apple TV+",
					StandalonePriceCents: 6900,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   56900,
				},
			},
			wantTotalSpent:   89800, // 32900 + 56900
			wantStandalone:   13800, // 6900 + 6900 (standalone counted per link)
			wantSavings:      -76000,
			wantOverlapCount: 1, // one overlap warning
		},
		{
			name:             "no links returns empty KPI",
			links:            []outbound.RecurringExpenseLink{},
			wantTotalSpent:   0,
			wantStandalone:   0,
			wantSavings:      0,
			wantOverlapCount: 0,
		},
		{
			name: "multiple overlaps detected",
			links: []outbound.RecurringExpenseLink{
				{
					RecurringExpenseID:   "exp-1",
					ServiceID:            "svc-1",
					ServiceName:          "Apple TV+",
					StandalonePriceCents: 6900,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   32900,
				},
				{
					RecurringExpenseID:   "exp-2",
					ServiceID:            "svc-1",
					ServiceName:          "Apple TV+",
					StandalonePriceCents: 6900,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   56900,
				},
				{
					RecurringExpenseID:   "exp-1",
					ServiceID:            "svc-2",
					ServiceName:          "Spotify",
					StandalonePriceCents: 12900,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   32900,
				},
				{
					RecurringExpenseID:   "exp-3",
					ServiceID:            "svc-2",
					ServiceName:          "Spotify",
					StandalonePriceCents: 12900,
					CustomPriceCents:     nil,
					ExpenseAmountCents:   19900,
				},
			},
			wantTotalSpent:   109700, // 32900 + 56900 + 19900 (3 expenses)
			wantStandalone:   39600,  // 6900 + 6900 + 12900 + 12900
			wantSavings:      -70100,
			wantOverlapCount: 2, // two overlap warnings
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockSubscriptionCatalogRepo{links: tt.links}
			uc := NewGetSubscriptionKPIUseCase(mockRepo)

			result, err := uc.Execute(context.Background(), "hh-1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.TotalSpentCents != tt.wantTotalSpent {
				t.Errorf("total_spent_cents = %d; want %d", result.TotalSpentCents, tt.wantTotalSpent)
			}

			if result.StandaloneTotalCents != tt.wantStandalone {
				t.Errorf("standalone_total_cents = %d; want %d", result.StandaloneTotalCents, tt.wantStandalone)
			}

			if result.NetSavingsCents != tt.wantSavings {
				t.Errorf("net_savings_cents = %d; want %d", result.NetSavingsCents, tt.wantSavings)
			}

			if len(result.OverlapWarnings) != tt.wantOverlapCount {
				t.Errorf("overlap_warnings count = %d; want %d", len(result.OverlapWarnings), tt.wantOverlapCount)
			}

			// Verify overlap warnings have correct structure
			for _, warning := range result.OverlapWarnings {
				if warning.CatalogServiceID == "" {
					t.Error("overlap warning missing catalog_service_id")
				}
				if warning.ServiceName == "" {
					t.Error("overlap warning missing service_name")
				}
				if len(warning.RecurringExpenseIDs) < 2 {
					t.Errorf("overlap warning for %q has %d expense IDs; want at least 2",
						warning.ServiceName, len(warning.RecurringExpenseIDs))
				}
			}
		})
	}
}

func int64Ptr(v int64) *int64 { return &v }
