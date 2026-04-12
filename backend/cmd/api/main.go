package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "micha/backend/internal/adapters/http"
	"micha/backend/internal/adapters/postgres"
	authapp "micha/backend/internal/application/auth"
	cardapp "micha/backend/internal/application/card"
	categoryapp "micha/backend/internal/application/category"
	expenseapp "micha/backend/internal/application/expense"
	householdapp "micha/backend/internal/application/household"
	memberapp "micha/backend/internal/application/member"
	recurringexpenseapp "micha/backend/internal/application/recurringexpense"
	settlementapp "micha/backend/internal/application/settlement"
	infraauth "micha/backend/internal/infrastructure/auth"
	"micha/backend/internal/infrastructure/config"
	"micha/backend/internal/infrastructure/migrations"
)

// uuidGenerator implements expenseapp.IDGenerator using UUID v4.
type uuidGenerator struct{}

func (uuidGenerator) NewID() string { return uuid.NewString() }

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("database unreachable", "error", err)
		os.Exit(1)
	}

	migrationsDir, err := resolveMigrationsDir()
	if err != nil {
		slog.Error("failed to locate migrations directory", "error", err)
		os.Exit(1)
	}

	if err := migrations.Apply(ctx, pool, migrationsDir); err != nil {
		slog.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}

	expenseRepo := postgres.NewExpenseRepository(pool)
	installmentRepo := postgres.NewInstallmentRepository(pool)
	recurringExpenseRepo := postgres.NewRecurringExpenseRepository(pool)
	householdRepo := postgres.NewHouseholdRepository(pool)
	memberRepo := postgres.NewMemberRepository(pool)
	memberInvitationRepo := postgres.NewMemberInvitationRepository(pool)
	cardRepo := postgres.NewCardRepository(pool)
	categoryRepo := postgres.NewCategoryRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	idGen := uuidGenerator{}

	hasher := infraauth.NewBcryptHasher()
	inviteSender := infraauth.NewLogInviteCodeSender()
	signer, err := infraauth.NewJWTSigner(cfg.JWTSecret)
	if err != nil {
		slog.Error("failed to create JWT signer", "error", err)
		os.Exit(1)
	}
	validator, err := infraauth.NewJWTValidator(cfg.JWTSecret)
	if err != nil {
		slog.Error("failed to create JWT validator", "error", err)
		os.Exit(1)
	}

	// Auth use cases and handler dependencies.
	authDeps := httpadapter.AuthHandlerDeps{
		Register: authapp.NewRegisterUserUseCase(userRepo, idGen, hasher),
		Login:    authapp.NewLoginUseCase(userRepo, hasher, signer),
		Members:  memberRepo,
	}

	// Expense use cases and handler dependencies.
	expenseDeps := httpadapter.ExpenseHandlerDeps{
		Register: expenseapp.NewRegisterExpenseUseCaseWithPolicy(expenseRepo, householdRepo, memberRepo, cardRepo, categoryRepo, installmentRepo, idGen, cfg.AllowOwnerOnBehalf),
		Get:      expenseapp.NewGetExpenseUseCase(expenseRepo),
		List:     expenseapp.NewListExpensesUseCase(expenseRepo),
		Patch:    expenseapp.NewPatchExpenseUseCase(expenseRepo),
		Delete:   expenseapp.NewDeleteExpenseUseCase(expenseRepo),
	}

	// Recurring expense use cases and handler dependencies.
	recurringExpenseDeps := httpadapter.RecurringExpenseHandlerDeps{
		Create:   recurringexpenseapp.NewCreateRecurringExpenseUseCase(recurringExpenseRepo, householdRepo, memberRepo, categoryRepo, idGen),
		Get:      recurringexpenseapp.NewGetRecurringExpenseUseCase(recurringExpenseRepo),
		List:     recurringexpenseapp.NewListRecurringExpensesUseCase(recurringExpenseRepo),
		Update:   recurringexpenseapp.NewUpdateRecurringExpenseUseCase(recurringExpenseRepo, categoryRepo),
		Delete:   recurringexpenseapp.NewDeleteRecurringExpenseUseCase(recurringExpenseRepo),
		Generate: recurringexpenseapp.NewGenerateRecurringExpensesUseCase(recurringExpenseRepo, expenseRepo, idGen),
	}

	// Household use cases and handler dependencies.
	householdDeps := httpadapter.HouseholdHandlerDeps{
		Register: householdapp.NewRegisterHouseholdUseCase(householdRepo, categoryRepo, idGen),
		List:     householdapp.NewListHouseholdsUseCase(householdRepo),
		Get:      householdapp.NewGetHouseholdUseCase(householdRepo),
		Update:   householdapp.NewUpdateHouseholdUseCase(householdRepo),
	}

	// Category use cases and handler dependencies.
	categoryDeps := httpadapter.CategoryHandlerDeps{
		Create: categoryapp.NewCreateCategoryUseCase(categoryRepo, idGen),
		List:   categoryapp.NewListCategoriesUseCase(categoryRepo),
		Delete: categoryapp.NewDeleteCategoryUseCase(categoryRepo),
	}

	// Split config use case.
	splitConfigDeps := httpadapter.SplitConfigHandlerDeps{
		Update: householdapp.NewUpdateSplitConfigUseCase(householdRepo),
	}

	// Member use cases and handler dependencies.
	memberDeps := httpadapter.MemberHandlerDeps{
		Register: memberapp.NewRegisterMemberUseCaseWithInvites(memberRepo, idGen, memberInvitationRepo, inviteSender),
		List:     memberapp.NewListMembersUseCase(memberRepo),
		Update:   memberapp.NewUpdateMemberUseCase(memberRepo),
		Delete:   memberapp.NewDeleteMemberUseCase(memberRepo),
	}

	// Card use cases and handler dependencies.
	cardDeps := httpadapter.CardHandlerDeps{
		Register: cardapp.NewRegisterCardUseCase(cardRepo, idGen),
		List:     cardapp.NewListCardsUseCase(cardRepo),
		Delete:   cardapp.NewDeleteCardUseCase(cardRepo),
	}

	// Server dependencies grouped by resource.
	serverDeps := httpadapter.ServerDependencies{
		Auth:             authDeps,
		Expense:          expenseDeps,
		RecurringExpense: recurringExpenseDeps,
		Household:        householdDeps,
		Member:           memberDeps,
		Card:             cardDeps,
		Settlement: httpadapter.SettlementHandlerDeps{
			Calculate: settlementapp.NewCalculateSettlementUseCase(householdRepo, memberRepo, expenseRepo, installmentRepo, recurringExpenseRepo),
		},
		Category:       categoryDeps,
		SplitConfig:    splitConfigDeps,
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: cfg.AllowedOrigins,
	}

	srv := httpadapter.NewServer(cfg.HTTPPort, serverDeps)

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: srv.Handler(),
	}

	go func() {
		slog.Info("api listening", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func resolveMigrationsDir() (string, error) {
	candidates := []string{
		"migrations",
		"backend/migrations",
		"../migrations",
	}

	for _, dir := range candidates {
		info, err := os.Stat(dir)
		if err == nil && info.IsDir() {
			abs, absErr := filepath.Abs(dir)
			if absErr != nil {
				return "", absErr
			}
			return abs, nil
		}
	}

	return "", errors.New("migrations directory not found")
}
