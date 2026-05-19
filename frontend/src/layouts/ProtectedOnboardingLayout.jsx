import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useAppShell } from '../context/AppShellContext'
import { OnboardingLayout } from './OnboardingLayout'

const POST_ONBOARDING_PATHS = ['/onboarding/household', '/onboarding/cards', '/onboarding/fixed-expenses']

/**
 * ProtectedOnboardingLayout — auth guard for onboarding routes.
 *
 * Flow: register → /onboarding/household → /onboarding/member → / (dashboard)
 *
 * - Unauthenticated → redirect to /login
 * - No household → allow /onboarding/household creation
 * - Has household but NO member → redirect to /onboarding/member
 * - Has household AND member → redirect to / (onboarding complete)
 *
 * Post-onboarding management routes (/onboarding/cards, /onboarding/fixed-expenses)
 * are always allowed so users can reach them from RulesPage.
 */
export function ProtectedOnboardingLayout() {
    const { isAuthenticated } = useAuth()
    const { households, loadingHouseholds, members, loadingMembers } = useAppShell()
    const { pathname } = useLocation()

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

    // Allow post-onboarding management routes even when onboarding is complete
    if (POST_ONBOARDING_PATHS.includes(pathname)) {
        return <OnboardingLayout />
    }

    // Has household AND member → onboarding complete, redirect to dashboard
    return <Navigate to="/" replace />
}
