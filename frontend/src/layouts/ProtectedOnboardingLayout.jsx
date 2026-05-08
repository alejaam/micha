import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { OnboardingLayout } from './OnboardingLayout'

/**
 * ProtectedOnboardingLayout — auth guard for onboarding routes.
 * Redirects unauthenticated users to /login.
 */
export function ProtectedOnboardingLayout() {
    const { isAuthenticated } = useAuth()

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />
    }

    return <OnboardingLayout />
}
