import { useCallback, useMemo } from 'react'
import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom'
import { AuthLayout } from './layouts/AuthLayout'
import { AppLayout } from './layouts/AppLayout'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { DashboardPage } from './pages/DashboardPage'
import { OnboardingHouseholdPage } from './pages/OnboardingHouseholdPage'
import { OnboardingMemberPage } from './pages/OnboardingMemberPage'
import { useHouseholds } from './hooks/useHouseholds'
import { useAuth } from './context/AuthContext'
import { getHealth } from './api'
import { useEffect, useState } from 'react'

/**
 * AppShell — holds shared state (households, health) that the layout and
 * pages both need. Rendered only when the router is mounted.
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
        setHouseholds,
        loadHouseholds,
    } = useHouseholds({ isAuthenticated, handleProtectedError })

    const selectedHousehold = useMemo(
        () => households.find((h) => h.id === householdId) ?? null,
        [householdId, households],
    )

    const handleReload = useCallback(async () => {
        await loadHouseholds()
    }, [loadHouseholds])

    const router = createBrowserRouter([
        {
            path: '/',
            element: (
                <AppLayout
                    health={health}
                    householdId={householdId}
                    households={households}
                    onHouseholdChange={setHouseholdId}
                    onReload={handleReload}
                    isLoading={loadingHouseholds}
                />
            ),
            children: [
                {
                    index: true,
                    element: (
                        <DashboardPage
                            householdId={householdId}
                            selectedHousehold={selectedHousehold}
                            loadHouseholds={loadHouseholds}
                            setHouseholdId={setHouseholdId}
                        />
                    ),
                },
                {
                    path: 'onboarding/household',
                    element: (
                        <OnboardingHouseholdPage
                            loadHouseholds={loadHouseholds}
                            setHouseholdId={setHouseholdId}
                        />
                    ),
                },
                {
                    path: 'onboarding/member',
                    element: (
                        <OnboardingMemberPage
                            householdId={householdId}
                        />
                    ),
                },
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

    return <RouterProvider router={router} />
}

export default AppShell
