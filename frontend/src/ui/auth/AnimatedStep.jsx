import { motion } from 'framer-motion'

const slideVariants = {
    enter: (direction) => ({
        opacity: 0,
        x: direction === 'forward' ? 40 : -40,
    }),
    center: { opacity: 1, x: 0 },
    exit: (direction) => ({
        opacity: 0,
        x: direction === 'forward' ? -40 : 40,
    }),
}

const transition = {
    duration: 0.25,
    ease: [0.16, 1, 0.3, 1],
}

/**
 * AnimatedStep — motion.div wrapper keyed by pathname for AnimatePresence.
 *
 * @param {string} pathname
 * @param {'forward'|'backward'} direction
 * @param {React.ReactNode} children
 */
export function AnimatedStep({ pathname, direction = 'forward', children }) {
    return (
        <motion.div
            key={pathname}
            variants={slideVariants}
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
