import { useCallback, useEffect, useMemo, useState } from 'react'
import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom'
import { getHealth } from './api'
import { AppShellContext } from './context/AppShellContext'
import { useAuth } from './context/AuthContext'
import { useDashboardUxState } from './hooks/useDashboardUxState'
import { useHouseholds } from './hooks/useHouseholds'
import { AppLayout } from './layouts/AppLayout'
import { AuthLayout } from './layouts/AuthLayout'
import { BalancesPage } from './pages/BalancesPage'
import { DashboardPage } from './pages/DashboardPage'
import { ExpensesPage } from './pages/ExpensesPage'
import { InstallmentsPage } from './pages/InstallmentsPage'
import { LoginPage } from './pages/LoginPage'
import { OnboardingCardsPage } from './pages/OnboardingCardsPage'
import { OnboardingFixedExpensesPage } from './pages/OnboardingFixedExpensesPage'
import { OnboardingHouseholdPage } from './pages/OnboardingHouseholdPage'
import { OnboardingMemberPage } from './pages/OnboardingMemberPage'
import { RegisterPage } from './pages/RegisterPage'
import { RulesPage } from './pages/RulesPage'

/**
 * Static router — created once at module level so React never tears down and
 * recreates the router tree on state changes inside AppShell.
 * Shared state is distributed via AppShellContext instead of router-element props.
 */
const router = createBrowserRouter([
    {
        path: '/',
        element: <AppLayout />,
        children: [
            { index: true, element: <DashboardPage /> },
            { path: 'expenses', element: <ExpensesPage /> },
            { path: 'balances', element: <BalancesPage /> },
            { path: 'installments', element: <InstallmentsPage /> },
            { path: 'rules', element: <RulesPage /> },
            { path: 'dashboard', element: <Navigate to="/" replace /> },
            { path: 'movements', element: <Navigate to="/expenses" replace /> },
            { path: 'settings', element: <Navigate to="/rules" replace /> },
            { path: 'onboarding/household', element: <OnboardingHouseholdPage /> },
            { path: 'onboarding/member', element: <OnboardingMemberPage /> },
            { path: 'onboarding/cards', element: <OnboardingCardsPage /> },
            { path: 'onboarding/fixed-expenses', element: <OnboardingFixedExpensesPage /> },
            { path: 'members/new', element: <OnboardingMemberPage /> },
        ],
    },
    {
        element: <AuthLayout />,
        children: [
            { path: '/login', element: <LoginPage /> },
            { path: '/register', element: <RegisterPage /> },
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
        periodStatus,
        setPeriodStatus,
        isMutationLocked,
    } = useDashboardUxState('open')

    const selectedHousehold = useMemo(
        () => households.find((h) => h.id === householdId) ?? null,
        [householdId, households],
    )

    const handleReload = useCallback(async () => {
        await loadHouseholds()
    }, [loadHouseholds])

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
    ])

    return (
        <AppShellContext.Provider value={shellValue}>
            <RouterProvider router={router} />
        </AppShellContext.Provider>
    )
}

export default AppShell
