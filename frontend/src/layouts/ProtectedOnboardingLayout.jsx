import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useAppShell } from '../context/AppShellContext'
import { OnboardingLayout } from './OnboardingLayout'

const POST_ONBOARDING_PATHS = ['/onboarding/household', '/onboarding/cards', '/onboarding/fixed-expenses']

/**
 * ProtectedOnboardingLayout — auth guard for onboarding routes.
 *
 * Flow: register → /onboarding/household → /onboarding/cards → /onboarding/fixed-expenses → / (dashboard)
 *
 * The owner member and first period are auto-created during household registration,
 * so there is no separate member-onboarding step.
 *
 * - Unauthenticated → redirect to /login
 * - No household → allow /onboarding/household creation
 * - Has household → allow onboarding routes (cards, fixed-expenses) or redirect to dashboard
 *
 * NOTE: We only block rendering on the FIRST load (loadingHouseholds && households.length === 0).
 * Subsequent background refetches (e.g. on focus) do NOT unmount the form to avoid losing user input.
 */
export function ProtectedOnboardingLayout() {
    const { isAuthenticated } = useAuth()
    const { households, loadingHouseholds } = useAppShell()
    const { pathname } = useLocation()

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />
    }

    // Only block on the very first load — subsequent refetches are silent and
    // we should keep showing the layout so the user doesn't lose form input.
    if (loadingHouseholds && households.length === 0) {
        return null
    }

    // No household → allow household creation onboarding
    if (households.length === 0) {
        return <OnboardingLayout />
    }

    // Allow post-onboarding management routes even when onboarding is complete
    if (POST_ONBOARDING_PATHS.includes(pathname)) {
        return <OnboardingLayout />
    }

    // Has household → onboarding complete, redirect to dashboard
    return <Navigate to="/" replace />
}
