import { buildRibbonState } from '../hooks/useDashboardUxState'

/**
 * PeriodStatusRibbon — compact status strip for the current period lifecycle.
 */
export function PeriodStatusRibbon({ status = 'open' }) {
    const ribbonState = buildRibbonState(status)
    const { status: normalizedStatus, stateLabel, description } = ribbonState

    return (
        <div
            className={`periodStatusRibbon periodStatusRibbon-${normalizedStatus}`}
            role="status"
            aria-live="polite"
            aria-label={`Period status: ${description}`}
        >
            <span className="periodStatusRibbonLabel">PERIOD</span>
            <span className="periodStatusRibbonState">{stateLabel}</span>
            <span className="periodStatusRibbonText">{description}</span>
        </div>
    )
}
