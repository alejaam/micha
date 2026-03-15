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
	categoryapp "micha/backend/internal/application/category"
	expenseapp "micha/backend/internal/application/expense"
	householdapp "micha/backend/internal/application/household"
	memberapp "micha/backend/internal/application/member"
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
	householdRepo := postgres.NewHouseholdRepository(pool)
	memberRepo := postgres.NewMemberRepository(pool)
	categoryRepo := postgres.NewCategoryRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	idGen := uuidGenerator{}

	hasher := infraauth.NewBcryptHasher()
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
	}

	// Expense use cases and handler dependencies.
	expenseDeps := httpadapter.ExpenseHandlerDeps{
		Register: expenseapp.NewRegisterExpenseUseCase(expenseRepo, householdRepo, memberRepo, idGen),
		Get:      expenseapp.NewGetExpenseUseCase(expenseRepo),
		List:     expenseapp.NewListExpensesUseCase(expenseRepo),
		Patch:    expenseapp.NewPatchExpenseUseCase(expenseRepo),
		Delete:   expenseapp.NewDeleteExpenseUseCase(expenseRepo),
	}

	// Household use cases and handler dependencies.
	householdDeps := httpadapter.HouseholdHandlerDeps{
		Register: householdapp.NewRegisterHouseholdUseCase(householdRepo, idGen),
		List:     householdapp.NewListHouseholdsUseCase(householdRepo),
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
		Register: memberapp.NewRegisterMemberUseCase(memberRepo, idGen),
		List:     memberapp.NewListMembersUseCase(memberRepo),
	}

	// Server dependencies grouped by resource.
	serverDeps := httpadapter.ServerDependencies{
		Auth:      authDeps,
		Expense:   expenseDeps,
		Household: householdDeps,
		Member:    memberDeps,
		Settlement: httpadapter.SettlementHandlerDeps{
			Calculate: settlementapp.NewCalculateSettlementUseCase(householdRepo, memberRepo, expenseRepo),
		},
		Category:     categoryDeps,
		SplitConfig:  splitConfigDeps,
		JWTValidator: validator,
		MemberRepo:   memberRepo,
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
