import { useCallback, useEffect, useMemo, useState } from 'react'
import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom'
import { getHealth } from './api'
import { AppShellContext } from './context/AppShellContext'
import { useAuth } from './context/AuthContext'
import { useDashboardUxState } from './hooks/useDashboardUxState'
import { useHouseholds } from './hooks/useHouseholds'
import { useMembers } from './hooks/useMembers'
import { AppLayout } from './layouts/AppLayout'
import { AuthLayout } from './layouts/AuthLayout'
import { ProtectedOnboardingLayout } from './layouts/ProtectedOnboardingLayout'
import { BalancesPage } from './pages/BalancesPage'
import { DashboardPage } from './pages/DashboardPage'
import { ExpensesPage } from './pages/ExpensesPage'
import { FixedExpensesPage } from './pages/FixedExpensesPage'
import { InstallmentsPage } from './pages/InstallmentsPage'
import { LoginPage } from './pages/LoginPage'
import { CardsPage } from './pages/CardsPage'
import { OnboardingCardsPage } from './pages/OnboardingCardsPage'
import { OnboardingFixedExpensesPage } from './pages/OnboardingFixedExpensesPage'
import { OnboardingHouseholdPage } from './pages/OnboardingHouseholdPage'
import { OnboardingMemberPage } from './pages/OnboardingMemberPage'
import { RegisterPage } from './pages/RegisterPage'
import { RulesPage } from './pages/RulesPage'

import { ErrorBoundary } from './components/ErrorBoundary'
import { ErrorPage } from './pages/ErrorPage'

/**
 * Static router — created once at module level so React never tears down and
 * recreates the router tree on state changes inside AppShell.
 * Shared state is distributed via AppShellContext instead of router-element props.
 */
const router = createBrowserRouter([
    {
        path: '/',
        element: <AppLayout />,
        errorElement: <ErrorPage />,
        children: [
            { index: true, element: <DashboardPage /> },
            { path: 'expenses', element: <ExpensesPage /> },
            { path: 'balances', element: <BalancesPage /> },
            { path: 'installments', element: <InstallmentsPage /> },
            { path: 'fixed-expenses', element: <FixedExpensesPage /> },
            { path: 'cards', element: <CardsPage /> },
            { path: 'settings/household', element: <OnboardingHouseholdPage /> },
            { path: 'rules', element: <RulesPage /> },
            { path: 'dashboard', element: <Navigate to="/" replace /> },
            { path: 'movements', element: <Navigate to="/expenses" replace /> },
            { path: 'settings', element: <Navigate to="/rules" replace /> },
            { path: 'onboarding/member', element: <OnboardingMemberPage /> },
            { path: 'members/new', element: <OnboardingMemberPage /> },
        ],
    },
    {
        element: <AuthLayout />,
        errorElement: <ErrorPage />,
        children: [
            { path: '/login', element: <LoginPage /> },
            { path: '/register', element: <RegisterPage /> },
        ],
    },
    {
        element: <ProtectedOnboardingLayout />,
        errorElement: <ErrorPage />,
        children: [
            { path: '/onboarding/household', element: <OnboardingHouseholdPage /> },
            { path: '/onboarding/cards', element: <OnboardingCardsPage /> },
            { path: '/onboarding/fixed-expenses', element: <OnboardingFixedExpensesPage /> },
        ],
    },
    { path: '*', element: <Navigate to="/" replace /> },
])

/**
 * AppShell — holds shared state (households, health) and provides it through
 * AppShellContext so layouts and pages can consume it without prop-drilling.
 */
function AppShell() {
    const { isAuthenticated, handleProtectedError } = useAuth()
    const [health, setHealth] = useState('checking...')

    useEffect(() => {
        let active = true
        getHealth()
            .then((status) => { if (active) setHealth(status === 'ok' ? 'ok' : status) })
            .catch(() => { if (active) setHealth('offline') })
        return () => { active = false }
    }, [])

    const {
        householdId,
        households,
        loadingHouseholds,
        setHouseholdId,
        loadHouseholds,
    } = useHouseholds({ isAuthenticated, handleProtectedError })

    const {
        members,
        loadingMembers,
        loadMembers,
    } = useMembers({ isAuthenticated, householdId, handleProtectedError })

    const {
        currentPeriod,
        periodStatus,
        setPeriodStatus,
        isMutationLocked,
        loadPeriod: reloadPeriod,
        selectedPeriodId,
        setSelectedPeriodId,
        consensus,
        consensusLoading,
        loadConsensus,
        isLoadingPeriod,
    } = useDashboardUxState(householdId)

    const selectedHousehold = useMemo(
        () => households.find((h) => h.id === householdId) ?? null,
        [householdId, households],
    )

    const handleReload = useCallback(async () => {
        await Promise.all([
            loadHouseholds(),
            reloadPeriod(),
        ])
    }, [loadHouseholds, reloadPeriod])

    const shellValue = useMemo(() => ({
        health,
        householdId,
        households,
        loadingHouseholds,
        selectedHousehold,
        setHouseholdId,
        loadHouseholds,
        handleReload,
        periodStatus,
        setPeriodStatus,
        isMutationLocked,
        currentPeriod,
        reloadPeriod,
        members,
        loadingMembers,
        loadMembers,
        selectedPeriodId,
        setSelectedPeriodId,
        consensus,
        consensusLoading,
        loadConsensus,
        isLoadingPeriod,
    }), [
        health,
        householdId,
        households,
        loadingHouseholds,
        selectedHousehold,
        setHouseholdId,
        loadHouseholds,
        handleReload,
        periodStatus,
        setPeriodStatus,
        isMutationLocked,
        currentPeriod,
        reloadPeriod,
        members,
        loadingMembers,
        loadMembers,
        selectedPeriodId,
        setSelectedPeriodId,
        consensus,
        consensusLoading,
        loadConsensus,
        isLoadingPeriod,
    ])

    return (
        <ErrorBoundary>
            <AppShellContext.Provider value={shellValue}>
                <RouterProvider router={router} />
            </AppShellContext.Provider>
        </ErrorBoundary>
    )
}

export default AppShell
