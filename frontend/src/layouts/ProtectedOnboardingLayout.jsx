import { Navigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useAppShell } from '../context/AppShellContext'
import { OnboardingLayout } from './OnboardingLayout'

/**
 * ProtectedOnboardingLayout — auth guard for onboarding routes.
 *
 * Flow: register → /onboarding/household → /onboarding/member → / (dashboard)
 *
 * - Unauthenticated → redirect to /login
 * - No household → allow /onboarding/household creation
 * - Has household but NO member → redirect to /onboarding/member
 * - Has household AND member → redirect to / (onboarding complete)
 */
export function ProtectedOnboardingLayout() {
    const { isAuthenticated } = useAuth()
    const { households, loadingHouseholds, members, loadingMembers } = useAppShell()

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />
    }

    if (loadingHouseholds) {
        return null
    }

    // No household → allow household creation onboarding
    if (households.length === 0) {
        return <OnboardingLayout />
    }

    // Has household — wait for members to load before redirecting
    if (loadingMembers) {
        return null
    }

    // Has household but no member → redirect to create the first member
    if (members.length === 0) {
        return <Navigate to="/onboarding/member" replace />
    }

    // Has household AND member → onboarding complete, redirect to dashboard
    return <Navigate to="/" replace />
}
