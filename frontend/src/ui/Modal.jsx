import { useEffect } from 'react'
import { motion } from 'framer-motion'

/**
 * Modal — reusable accessible modal overlay with animations.
 *
 * @param {string} title
 * @param {()=>void} onClose
 * @param {React.ReactNode} children
 */
export function Modal({ title, onClose, children }) {
    // Close on Escape
    useEffect(() => {
        function handleKey(e) {
            if (e.key === 'Escape') onClose()
        }
        document.addEventListener('keydown', handleKey)
        return () => document.removeEventListener('keydown', handleKey)
    }, [onClose])

    // Prevent background scroll
    useEffect(() => {
        const originalHtmlOverflow = document.documentElement.style.overflow
        const originalBodyOverflow = document.body.style.overflow

        document.documentElement.style.overflow = 'hidden'
        document.body.style.overflow = 'hidden'

        return () => {
            document.documentElement.style.overflow = originalHtmlOverflow
            document.body.style.overflow = originalBodyOverflow
        }
    }, [])

    return (
        <motion.div
            className="modalOverlay"
            role="dialog"
            aria-modal="true"
            aria-label={title}
            onClick={onClose}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
        >
            <motion.div
                className="modalPanel card"
                onClick={(e) => e.stopPropagation()}
                initial={{ y: 20, opacity: 0, scale: 0.95 }}
                animate={{ y: 0, opacity: 1, scale: 1 }}
                exit={{ y: 20, opacity: 0, scale: 0.95 }}
                transition={{ type: 'spring', damping: 25, stiffness: 300 }}
            >
                <div className="modalHeader">
                    <h2 className="modalTitle">{title}</h2>
                    <button
                        type="button"
                        className="btn btnGhost btnSm btnIcon"
                        onClick={onClose}
                        aria-label="Close"
                    >
                        ✕
                    </button>
                </div>
                {children}
            </motion.div>
        </motion.div>
    )
}
