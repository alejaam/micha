import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { AppHeader } from '../components/AppHeader'
import { BottomNav } from '../components/BottomNav'
import { PrimaryNav } from '../components/PrimaryNav'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'

/**
 * AppLayout — wraps protected routes.
 * Redirects unauthenticated users to /login.
 * Provides the global header, reading shared state from AppShellContext.
 */
export function AppLayout() {
    const { isAuthenticated, logout } = useAuth()
    const location = useLocation()
    const {
        health,
        householdId,
        households,
        setHouseholdId,
        handleReload,
        loadingHouseholds,
        periodStatus,
        isMutationLocked,
    } = useAppShell()

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />
    }

    const isMobile = typeof window !== 'undefined' && window.innerWidth < 880
    const hidePrimaryNav = location.pathname.startsWith('/onboarding') || location.pathname.startsWith('/members/new') || !isMobile

    return (
        <div className="page">
            <AppHeader
                health={health}
                householdId={householdId}
                households={households}
                onHouseholdChange={setHouseholdId}
                onReload={handleReload}
                onLogout={logout}
                isLoading={loadingHouseholds}
                periodStatus={periodStatus}
                isMutationLocked={isMutationLocked}
            />
            {(!hidePrimaryNav && isMobile) && <PrimaryNav />}
            <Outlet />
            <BottomNav />
        </div>
    )
}
