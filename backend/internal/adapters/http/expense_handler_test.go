package httpadapter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"

	httpadapter "micha/backend/internal/adapters/http"
)

// --- POST /v1/expenses ---

func TestExpenseHandler_Create_Success(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{
				returnOutput: inbound.RegisterExpenseOutput{ExpenseID: "new-expense-id"},
			},
			Get:    &mockGetExpense{},
			List:   &mockListExpenses{},
			Patch:  &mockPatchExpense{},
			Delete: &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"household_id":      "household-1",
		"paid_by_member_id": "member-1",
		"amount_cents":      15000,
		"description":       "Groceries",
		"is_shared":         true,
		"currency":          "MXN",
		"payment_method":    "card",
		"expense_type":      "variable",
		"card_name":         "BBVA Azul",
		"category":          "food",
	}

	req := makeJSONRequest(t, "POST", "/v1/expenses", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusCreated)
	}

	resp := parseJSONResponse(t, rec)
	data, _ := resp["data"].(map[string]any)
	if data["expense_id"] != "new-expense-id" {
		t.Errorf("expense_id = %v; want new-expense-id", data["expense_id"])
	}
}

func TestExpenseHandler_Create_DefaultCurrency(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{
				returnOutput: inbound.RegisterExpenseOutput{ExpenseID: "e1"},
			},
			Get:    &mockGetExpense{},
			List:   &mockListExpenses{},
			Patch:  &mockPatchExpense{},
			Delete: &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"household_id":      "household-1",
		"paid_by_member_id": "member-1",
		"amount_cents":      15000,
		"description":       "Groceries",
		"payment_method":    "card",
		"expense_type":      "variable",
		"category":          "food",
	}

	req := makeJSONRequest(t, "POST", "/v1/expenses", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusCreated)
	}
}

func TestExpenseHandler_Create_PropagatesCurrentUserID(t *testing.T) {
	t.Parallel()

	register := &mockRegisterExpense{
		returnOutput: inbound.RegisterExpenseOutput{ExpenseID: "e-context"},
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: register,
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-admin-1", returnEmail: "admin@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"household_id":      "household-1",
		"paid_by_member_id": "member-1",
		"amount_cents":      11000,
		"description":       "Internet",
		"is_shared":         true,
		"currency":          "MXN",
		"payment_method":    "transfer",
		"expense_type":      "fixed",
		"category":          "internet",
	}

	req := makeJSONRequest(t, "POST", "/v1/expenses", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusCreated)
	}

	if register.lastInput.CurrentUserID != "user-admin-1" {
		t.Errorf("current_user_id = %q; want %q", register.lastInput.CurrentUserID, "user-admin-1")
	}
}

func TestExpenseHandler_Create_Forbidden(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{returnErr: shared.ErrForbidden},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-member-2", returnEmail: "user2@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"household_id":      "household-1",
		"paid_by_member_id": "member-1",
		"amount_cents":      15000,
		"description":       "Unauthorized try",
		"payment_method":    "card",
		"expense_type":      "variable",
		"category":          "food",
	}

	req := makeJSONRequest(t, "POST", "/v1/expenses", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusForbidden)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "FORBIDDEN" {
		t.Errorf("code = %v; want FORBIDDEN", errObj["code"])
	}
}

func TestExpenseHandler_Create_InvalidMoney(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{returnErr: shared.ErrInvalidMoney},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"household_id":      "household-1",
		"paid_by_member_id": "member-1",
		"amount_cents":      0,
		"description":       "Invalid",
		"payment_method":    "card",
		"expense_type":      "variable",
		"category":          "food",
	}

	req := makeJSONRequest(t, "POST", "/v1/expenses", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "INVALID_MONEY" {
		t.Errorf("code = %v; want INVALID_MONEY", errObj["code"])
	}
}

func TestExpenseHandler_Create_RejectsNegativeAmount(t *testing.T) {
	t.Parallel()

	register := &mockRegisterExpense{returnOutput: inbound.RegisterExpenseOutput{ExpenseID: "must-not-be-used"}}
	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: register,
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"household_id":      "household-1",
		"paid_by_member_id": "member-1",
		"amount_cents":      -10,
		"description":       "Invalid",
		"payment_method":    "card",
		"expense_type":      "variable",
		"category":          "food",
	}

	req := makeJSONRequest(t, "POST", "/v1/expenses", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "INVALID_MONEY" {
		t.Errorf("code = %v; want INVALID_MONEY", errObj["code"])
	}

	if register.lastInput.HouseholdID != "" {
		t.Errorf("register use case should not be called when amount is invalid")
	}
}

func TestExpenseHandler_Create_InvalidPaymentMethod(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{returnErr: expense.ErrInvalidPaymentMethod},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"household_id":      "household-1",
		"paid_by_member_id": "member-1",
		"amount_cents":      15000,
		"description":       "Test",
		"payment_method":    "invalid",
		"expense_type":      "variable",
		"category":          "food",
	}

	req := makeJSONRequest(t, "POST", "/v1/expenses", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "INVALID_PAYMENT_METHOD" {
		t.Errorf("code = %v; want INVALID_PAYMENT_METHOD", errObj["code"])
	}
}

// --- GET /v1/expenses/{id} ---

func TestExpenseHandler_Get_Success(t *testing.T) {
	t.Parallel()

	e, _ := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:             expense.ID("550e8400-e29b-41d4-a716-446655440000"),
		HouseholdID:    "household-1",
		PaidByMemberID: "member-1",
		AmountCents:    15000,
		Description:    "Test expense",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  expense.PaymentMethodCard,
		ExpenseType:    expense.ExpenseTypeVariable,
		CardName:       "BBVA Azul",
		CategoryID:     "cat-food",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{returnExpense: e},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/expenses/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	resp := parseJSONResponse(t, rec)
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("data field missing or not an object")
	}
	if data["id"] != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("id = %v; want 550e8400-e29b-41d4-a716-446655440000", data["id"])
	}
}

func TestExpenseHandler_Get_NotFound(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{returnErr: shared.ErrNotFound},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/expenses/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusNotFound)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "NOT_FOUND" {
		t.Errorf("code = %v; want NOT_FOUND", errObj["code"])
	}
}

func TestExpenseHandler_Get_InvalidID(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/expenses/not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "INVALID_ID" {
		t.Errorf("code = %v; want INVALID_ID", errObj["code"])
	}
}

// --- GET /v1/expenses?household_id=... ---

func TestExpenseHandler_List_Success(t *testing.T) {
	t.Parallel()

	e1, _ := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:             expense.ID("e1"),
		HouseholdID:    "household-1",
		PaidByMemberID: "member-1",
		AmountCents:    10000,
		Description:    "Expense 1",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  expense.PaymentMethodCard,
		ExpenseType:    expense.ExpenseTypeVariable,
		CategoryID:     "cat-food",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	e2, _ := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:             expense.ID("e2"),
		HouseholdID:    "household-1",
		PaidByMemberID: "member-2",
		AmountCents:    20000,
		Description:    "Expense 2",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  expense.PaymentMethodCash,
		ExpenseType:    expense.ExpenseTypeFixed,
		CategoryID:     "cat-rent",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{returnExpenses: []expense.Expense{e1, e2}},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/expenses?household_id=household-1&limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	resp := parseJSONResponse(t, rec)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatal("data field missing or not an array")
	}
	if len(data) != 2 {
		t.Errorf("len(data) = %d; want 2", len(data))
	}
}

func TestExpenseHandler_List_MissingHouseholdID(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/expenses", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "VALIDATION_ERROR" {
		t.Errorf("code = %v; want VALIDATION_ERROR", errObj["code"])
	}
}

// --- PATCH /v1/expenses/{id} ---

func TestExpenseHandler_Patch_Success(t *testing.T) {
	t.Parallel()

	updated, _ := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:             expense.ID("550e8400-e29b-41d4-a716-446655440000"),
		HouseholdID:    "household-1",
		PaidByMemberID: "member-1",
		AmountCents:    25000,
		Description:    "Updated description",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  expense.PaymentMethodCard,
		ExpenseType:    expense.ExpenseTypeVariable,
		CategoryID:     "cat-food",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{returnExpense: updated},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"description":  "Updated description",
		"amount_cents": 25000,
	}

	req := makeJSONRequest(t, "PATCH", "/v1/expenses/550e8400-e29b-41d4-a716-446655440000", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	resp := parseJSONResponse(t, rec)
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("data field missing or not an object")
	}
	if data["id"] != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("id = %v; want 550e8400-e29b-41d4-a716-446655440000", data["id"])
	}
}

func TestExpenseHandler_Patch_NotFound(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{returnErr: shared.ErrNotFound},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"description": "Updated",
	}

	req := makeJSONRequest(t, "PATCH", "/v1/expenses/550e8400-e29b-41d4-a716-446655440000", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusNotFound)
	}
}

func TestExpenseHandler_Patch_RejectsNonPositiveAmount(t *testing.T) {
	t.Parallel()

	patch := &mockPatchExpense{}
	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    patch,
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	body := map[string]any{
		"amount_cents": 0,
	}

	req := makeJSONRequest(t, "PATCH", "/v1/expenses/550e8400-e29b-41d4-a716-446655440000", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "INVALID_MONEY" {
		t.Errorf("code = %v; want INVALID_MONEY", errObj["code"])
	}

	if patch.lastInput.ID != "" {
		t.Errorf("patch use case should not be called when amount is invalid")
	}
}

// --- DELETE /v1/expenses/{id} ---

func TestExpenseHandler_Delete_Success(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "DELETE", "/v1/expenses/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusNoContent)
	}
}

func TestExpenseHandler_Delete_NotFound(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{returnErr: shared.ErrNotFound},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "DELETE", "/v1/expenses/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusNotFound)
	}
}

func TestExpenseHandler_Delete_AlreadyDeleted(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Expense: httpadapter.ExpenseHandlerDeps{
			Register: &mockRegisterExpense{},
			Get:      &mockGetExpense{},
			List:     &mockListExpenses{},
			Patch:    &mockPatchExpense{},
			Delete:   &mockDeleteExpense{returnErr: shared.ErrAlreadyDeleted},
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "DELETE", "/v1/expenses/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "ALREADY_DELETED" {
		t.Errorf("code = %v; want ALREADY_DELETED", errObj["code"])
	}
}
