import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useAppShell } from '../context/AppShellContext'
import { OnboardingLayout } from './OnboardingLayout'

/**
 * ProtectedOnboardingLayout — auth guard for onboarding routes.
 * Redirects unauthenticated users to /login.
 * Redirects users who already have a household away from /onboarding/household
 * to prevent accidental duplicate household creation.
 */
export function ProtectedOnboardingLayout() {
    const { isAuthenticated } = useAuth()
    const { households, loadingHouseholds } = useAppShell()
    const location = useLocation()

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />
    }

    if (loadingHouseholds) {
        return null
    }

    if (households.length > 0 && location.pathname === '/onboarding/household') {
        return <Navigate to="/" replace />
    }

    return <OnboardingLayout />
}
