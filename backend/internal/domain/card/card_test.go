package card_test

import (
	"testing"
	"time"

	"micha/backend/internal/domain/card"
	"micha/backend/internal/domain/shared"
)

func TestNewFromAttributes(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		attrs   card.Attributes
		wantErr error
	}{
		{
			name: "valid card",
			attrs: card.Attributes{
				ID:          "card-123",
				HouseholdID: "house-123",
				BankName:    "Nu",
				CardName:    "Nu Plata",
				CutoffDay:   15,
				CreatedAt:   now,
			},
			wantErr: nil,
		},
		{
			name: "missing id",
			attrs: card.Attributes{
				ID:          "",
				HouseholdID: "house-123",
				BankName:    "Nu",
				CardName:    "Nu Plata",
				CutoffDay:   15,
				CreatedAt:   now,
			},
			wantErr: shared.ErrInvalidID,
		},
		{
			name: "missing household id",
			attrs: card.Attributes{
				ID:          "card-123",
				HouseholdID: "",
				BankName:    "Nu",
				CardName:    "Nu Plata",
				CutoffDay:   15,
				CreatedAt:   now,
			},
			wantErr: shared.ErrInvalidID,
		},
		{
			name: "missing bank name",
			attrs: card.Attributes{
				ID:          "card-123",
				HouseholdID: "house-123",
				BankName:    "  ",
				CardName:    "Nu Plata",
				CutoffDay:   15,
				CreatedAt:   now,
			},
			wantErr: card.ErrInvalidBankName,
		},
		{
			name: "missing card name",
			attrs: card.Attributes{
				ID:          "card-123",
				HouseholdID: "house-123",
				BankName:    "Nu",
				CardName:    "",
				CutoffDay:   15,
				CreatedAt:   now,
			},
			wantErr: card.ErrInvalidCardName,
		},
		{
			name: "invalid cutoff day (under 1)",
			attrs: card.Attributes{
				ID:          "card-123",
				HouseholdID: "house-123",
				BankName:    "Nu",
				CardName:    "Nu Plata",
				CutoffDay:   0,
				CreatedAt:   now,
			},
			wantErr: card.ErrInvalidCutoffDay,
		},
		{
			name: "invalid cutoff day (over 31)",
			attrs: card.Attributes{
				ID:          "card-123",
				HouseholdID: "house-123",
				BankName:    "Nu",
				CardName:    "Nu Plata",
				CutoffDay:   32,
				CreatedAt:   now,
			},
			wantErr: card.ErrInvalidCutoffDay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := card.NewFromAttributes(tt.attrs)

			if err != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if result.ID() != tt.attrs.ID {
					t.Errorf("got %q, want %q", result.ID(), tt.attrs.ID)
				}
				if result.BankName() != tt.attrs.BankName {
					t.Errorf("got %q, want %q", result.BankName(), tt.attrs.BankName)
				}
				// Verify UpdatedAt inherits from CreatedAt when missing
				if tt.attrs.UpdatedAt.IsZero() && result.UpdatedAt() != tt.attrs.CreatedAt {
					t.Errorf("UpdatedAt not defaulting to CreatedAt properly")
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	now := time.Now()
	c, _ := card.New("1", "h1", "Nu", "Nu", 15, now)

	if c.IsDeleted() {
		t.Errorf("expected new card to not be deleted")
	}

	deleteTime := now.Add(1 * time.Hour)
	c = c.Delete(deleteTime)

	if !c.IsDeleted() {
		t.Errorf("expected card to be marked as deleted")
	}
	if c.DeletedAt() == nil || !c.DeletedAt().Equal(deleteTime) {
		t.Errorf("DeletedAt not set correctly")
	}
	if !c.UpdatedAt().Equal(deleteTime) {
		t.Errorf("UpdatedAt should be updated upon Delete")
	}
}
