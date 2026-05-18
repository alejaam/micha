import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { AnimatePresence } from 'framer-motion'
import { StepIndicator } from '../ui/auth'
import { useSlideDirection } from '../hooks/useSlideDirection'
import { useAuth } from '../context/AuthContext'

// eslint-disable-next-line react-refresh/only-export-components
export const ONBOARDING_STEPS = [
    { path: '/onboarding/household', label: 'Hogar' },
    { path: '/onboarding/cards', label: 'Tarjetas' },
    { path: '/onboarding/fixed-expenses', label: 'Gastos fijos' },
]

const STEP_PATHS = ONBOARDING_STEPS.map((s) => s.path)

/**
 * OnboardingLayout — step indicator, minimal header, AnimatePresence wrapper.
 */
export function OnboardingLayout() {
    const { pathname } = useLocation()
    const direction = useSlideDirection(STEP_PATHS)
    const { logout } = useAuth()
    const navigate = useNavigate()

    const handleLogout = () => {
        logout()
        navigate('/login', { replace: true })
    }

    return (
        <main className="onboardingShell">
            <div className="onboardingHeaderBar">
                <div className="brand">
                    <div className="brandIcon" aria-hidden>💸</div>
                    <div>
                        <div className="brandName">micha</div>
                    </div>
                </div>
                <button type="button" className="btn btnGhost btnSm" onClick={handleLogout}>
                    Cerrar sesión
                </button>
            </div>
            <StepIndicator steps={ONBOARDING_STEPS} currentPath={pathname} />
            <AnimatePresence mode="wait" custom={direction}>
                <Outlet key={pathname} />
            </AnimatePresence>
        </main>
    )
}
