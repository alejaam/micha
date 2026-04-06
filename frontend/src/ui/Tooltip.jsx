import { useState } from 'react'

/**
 * Tooltip — accessible inline tooltip with hover/focus trigger.
 * Shows helpful context for complex fields and panels.
 *
 * @param {string} text - The tooltip content
 * @param {string} [position='top'] - Tooltip position: top, bottom, left, right
 */
export function Tooltip({ text, position = 'top' }) {
    const [visible, setVisible] = useState(false)

    if (!text) return null

    return (
        <span
            className={`tooltipTrigger tooltip-${position}`}
            onMouseEnter={() => setVisible(true)}
            onMouseLeave={() => setVisible(false)}
            onFocus={() => setVisible(true)}
            onBlur={() => setVisible(false)}
            tabIndex={0}
            role="button"
            aria-label={`More info: ${text}`}
        >
            <span className="tooltipIcon" aria-hidden>?</span>
            {visible && (
                <span className="tooltipContent" role="tooltip">
                    {text}
                </span>
            )}
        </span>
    )
}