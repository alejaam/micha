import { useCallback, useEffect, useMemo, useState } from 'react'
import { createBrowserRouter, Navigate, RouterProvider } from 'react-router-dom'
import { getHealth } from './api'
import { AppShellContext } from './context/AppShellContext'
import { useAuth } from './context/AuthContext'
import { useHouseholds } from './hooks/useHouseholds'
import { AppLayout } from './layouts/AppLayout'
import { AuthLayout } from './layouts/AuthLayout'
import { DashboardPage } from './pages/DashboardPage'
import { LoginPage } from './pages/LoginPage'
import { OnboardingCardsPage } from './pages/OnboardingCardsPage'
import { OnboardingHouseholdPage } from './pages/OnboardingHouseholdPage'
import { OnboardingMemberPage } from './pages/OnboardingMemberPage'
import { RegisterPage } from './pages/RegisterPage'

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
            { path: 'onboarding/household', element: <OnboardingHouseholdPage /> },
            { path: 'onboarding/member', element: <OnboardingMemberPage /> },
            { path: 'onboarding/cards', element: <OnboardingCardsPage /> },
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
    }), [health, householdId, households, loadingHouseholds, selectedHousehold, setHouseholdId, loadHouseholds, handleReload])

    return (
        <AppShellContext.Provider value={shellValue}>
            <RouterProvider router={router} />
        </AppShellContext.Provider>
    )
}

export default AppShell
