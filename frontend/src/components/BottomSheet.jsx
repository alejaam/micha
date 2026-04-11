import { useEffect, useRef } from 'react'
import { AnimatePresence, motion } from 'framer-motion'

export function BottomSheet({
  open,
  title,
  onClose,
  children,
}) {
  const closeButtonRef = useRef(null)

  useEffect(() => {
    if (!open) return

    function onKeyDown(event) {
      if (event.key === 'Escape') onClose()
    }

    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', onKeyDown)
    closeButtonRef.current?.focus()

    return () => {
      document.body.style.overflow = previousOverflow
      document.removeEventListener('keydown', onKeyDown)
    }
  }, [open, onClose])

  return (
    <AnimatePresence>
      {open ? (
        <motion.div
          className="bottomSheetOverlay"
          onClick={onClose}
          role="presentation"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.18, ease: 'easeOut' }}
        >
          <motion.section
            className="bottomSheetPanel"
            onClick={(event) => event.stopPropagation()}
            role="dialog"
            aria-modal="true"
            aria-label={title}
            tabIndex={-1}
            initial={{ y: '100%' }}
            animate={{ y: 0 }}
            exit={{ y: '100%' }}
            transition={{ duration: 0.2, ease: 'easeOut' }}
          >
            <button
              ref={closeButtonRef}
              type="button"
              className="bottomSheetHandle"
              aria-label="Close panel"
              onClick={onClose}
            />
            <div className="bottomSheetHeader">
              <h2 className="bottomSheetTitle">{title}</h2>
              <button type="button" className="btn btnGhost btnSm btnIcon" onClick={onClose} aria-label="Close">
                ✕
              </button>
            </div>
            <div className="bottomSheetBody">{children}</div>
          </motion.section>
        </motion.div>
      ) : null}
    </AnimatePresence>
  )
}
