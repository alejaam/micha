import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { AppHeader } from '../components/AppHeader'

/**
 * AppLayout — wraps protected routes.
 * Redirects unauthenticated users to /login.
 * Provides the global header.
 */
export function AppLayout({ health, householdId, households, onHouseholdChange, onReload, isLoading }) {
    const { isAuthenticated, logout } = useAuth()

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />
    }

    return (
        <div className="page">
            <AppHeader
                health={health}
                householdId={householdId}
                households={households}
                onHouseholdChange={onHouseholdChange}
                onReload={onReload}
                onLogout={logout}
                isLoading={isLoading}
            />
            <Outlet />
        </div>
    )
}
