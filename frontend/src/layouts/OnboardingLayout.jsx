import { Outlet, useLocation } from 'react-router-dom'
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion'
import { StepIndicator } from '../ui/auth'
import { useSlideDirection } from '../hooks/useSlideDirection'

export const ONBOARDING_STEPS = [
    { path: '/onboarding/household', label: 'Hogar' },
    { path: '/onboarding/cards', label: 'Tarjetas' },
    { path: '/onboarding/fixed-expenses', label: 'Gastos fijos' },
]

const STEP_PATHS = ONBOARDING_STEPS.map((s) => s.path)

const variants = {
    enter: (dir) => ({
        opacity: 0,
        x: dir === 'forward' ? 40 : -40,
    }),
    center: { opacity: 1, x: 0 },
    exit: (dir) => ({
        opacity: 0,
        x: dir === 'forward' ? -40 : 40,
    }),
}

const transition = { duration: 0.25, ease: [0.16, 1, 0.3, 1] }

/**
 * OnboardingLayout — dark theme, step indicator, AnimatePresence wrapper.
 */
export function OnboardingLayout() {
    const { pathname } = useLocation()
    const direction = useSlideDirection(STEP_PATHS)
    const prefersReducedMotion = useReducedMotion()

    const reducedVariants = {
        enter: { opacity: 1, x: 0 },
        center: { opacity: 1, x: 0 },
        exit: { opacity: 1, x: 0 },
    }

    const reducedTransition = { duration: 0.01 }

    return (
        <main className="premiumDark onboardingShell">
            <StepIndicator steps={ONBOARDING_STEPS} currentPath={pathname} />
            <AnimatePresence mode="wait" custom={direction}>
                <motion.div
                    key={pathname}
                    custom={direction}
                    variants={prefersReducedMotion ? reducedVariants : variants}
                    initial="enter"
                    animate="center"
                    exit="exit"
                    transition={prefersReducedMotion ? reducedTransition : transition}
                >
                    <Outlet />
                </motion.div>
            </AnimatePresence>
        </main>
    )
}
