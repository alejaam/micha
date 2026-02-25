package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "micha/backend/internal/adapters/http"
	"micha/backend/internal/adapters/postgres"
	expenseapp "micha/backend/internal/application/expense"
	"micha/backend/internal/infrastructure/config"
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

	repo := postgres.NewExpenseRepository(pool)
	idGen := uuidGenerator{}

	// Expense use cases and handler dependencies.
	expenseDeps := httpadapter.ExpenseHandlerDeps{
		Register: expenseapp.NewRegisterExpenseUseCase(repo, idGen),
		Get:      expenseapp.NewGetExpenseUseCase(repo),
		List:     expenseapp.NewListExpensesUseCase(repo),
		Patch:    expenseapp.NewPatchExpenseUseCase(repo),
		Delete:   expenseapp.NewDeleteExpenseUseCase(repo),
	}

	// Server dependencies grouped by resource.
	serverDeps := httpadapter.ServerDependencies{
		Expense: expenseDeps,
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
