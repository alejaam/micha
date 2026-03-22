import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

/**
 * AuthLayout — wraps public routes (login, register).
 * Redirects authenticated users to the dashboard.
 */
export function AuthLayout() {
    const { isAuthenticated } = useAuth()

    if (isAuthenticated) {
        return <Navigate to="/" replace />
    }

    return (
        <main className="authShell">
            <Outlet />
        </main>
    )
}
