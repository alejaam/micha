import { motion, useReducedMotion } from 'framer-motion'

/**
 * AnimatedStep — motion.div wrapper keyed by pathname for AnimatePresence.
 * Respects prefers-reduced-motion by disabling all animations.
 *
 * @param {string} pathname
 * @param {'forward'|'backward'} direction
 * @param {React.ReactNode} children
 */
export function AnimatedStep({ pathname, direction = 'forward', children }) {
    const prefersReducedMotion = useReducedMotion()

    const variants = prefersReducedMotion
        ? {
              enter: { opacity: 1, x: 0 },
              center: { opacity: 1, x: 0 },
              exit: { opacity: 1, x: 0 },
          }
        : {
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

    const transition = prefersReducedMotion
        ? { duration: 0.01 }
        : { duration: 0.25, ease: [0.16, 1, 0.3, 1] }

    return (
        <motion.div
            key={pathname}
            variants={variants}
            initial="enter"
            animate="center"
            exit="exit"
            custom={direction}
            transition={transition}
        >
            {children}
        </motion.div>
    )
}
