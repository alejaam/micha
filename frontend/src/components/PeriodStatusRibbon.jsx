const PERIOD_STATUS_MAP = {
    open: {
        stateLabel: '[OPEN]',
        description: 'Period open — mutating actions are enabled.',
    },
    review: {
        stateLabel: '[REVIEW]',
        description: 'Period under review — mutating actions are temporarily locked.',
    },
    closed: {
        stateLabel: '[CLOSED]',
        description: 'Period closed — mutating actions are disabled.',
    },
}

/**
 * PeriodStatusRibbon — compact status strip for the current period lifecycle.
 */
export function PeriodStatusRibbon({ status = 'open' }) {
    const normalizedStatus = PERIOD_STATUS_MAP[status] ? status : 'open'
    const content = PERIOD_STATUS_MAP[normalizedStatus]

    return (
        <div
            className={`periodStatusRibbon periodStatusRibbon-${normalizedStatus}`}
            role="status"
            aria-live="polite"
            aria-label={`Period status: ${content.description}`}
        >
            <span className="periodStatusRibbonLabel">PERIOD</span>
            <span className="periodStatusRibbonState">{content.stateLabel}</span>
            <span className="periodStatusRibbonText">{content.description}</span>
        </div>
    )
}
