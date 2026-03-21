import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useAppShell } from '../context/AppShellContext'
import { AppHeader } from '../components/AppHeader'
import { BottomNav } from '../components/BottomNav'

/**
 * AppLayout — wraps protected routes.
 * Redirects unauthenticated users to /login.
 * Provides the global header, reading shared state from AppShellContext.
 */
export function AppLayout() {
    const { isAuthenticated, logout } = useAuth()
    const { health, householdId, households, setHouseholdId, handleReload, loadingHouseholds } = useAppShell()

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />
    }

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
            />
            <Outlet />
            <BottomNav />
        </div>
    )
}
