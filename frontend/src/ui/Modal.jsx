import { useEffect } from 'react'

/**
 * Modal — reusable accessible modal overlay.
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

    // Prevent body scroll
    useEffect(() => {
        const prev = document.body.style.overflow
        document.body.style.overflow = 'hidden'
        return () => { document.body.style.overflow = prev }
    }, [])

    return (
        <div className="modalOverlay" role="dialog" aria-modal="true" aria-label={title} onClick={onClose}>
            <div className="modalPanel card" onClick={(e) => e.stopPropagation()}>
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
            </div>
        </div>
    )
}
