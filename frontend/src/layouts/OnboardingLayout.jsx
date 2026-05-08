import { Outlet, useLocation } from 'react-router-dom'
import { AnimatePresence } from 'framer-motion'
import { StepIndicator } from '../ui/auth'
import { useSlideDirection } from '../hooks/useSlideDirection'

export const ONBOARDING_STEPS = [
    { path: '/onboarding/household', label: 'Hogar' },
    { path: '/onboarding/cards', label: 'Tarjetas' },
    { path: '/onboarding/fixed-expenses', label: 'Gastos fijos' },
]

const STEP_PATHS = ONBOARDING_STEPS.map((s) => s.path)

/**
 * OnboardingLayout — dark theme, step indicator, AnimatePresence wrapper.
 */
export function OnboardingLayout() {
    const { pathname } = useLocation()
    const direction = useSlideDirection(STEP_PATHS)

    return (
        <main className="premiumDark onboardingShell">
            <StepIndicator steps={ONBOARDING_STEPS} currentPath={pathname} />
            <AnimatePresence mode="wait" custom={direction}>
                <Outlet key={pathname} />
            </AnimatePresence>
        </main>
    )
}
