import { formatCurrency } from '../utils'

/**
 * PeriodBar — redesigned PeriodStatusRibbon that shows period title, date range,
 * balance, and status as a full-width bar at the top of the dashboard.
 *
 * @param {object} currentPeriod - { startDate, endDate, status }
 * @param {number} balance       - positive = debt (total spent), negative = owed (overpaid surplus)
 */
export function PeriodBar({ currentPeriod, balance = 0 }) {
    if (!currentPeriod) return null
    const { startDate, endDate, status = 'open' } = currentPeriod

    const formatDate = (iso) => {
        if (!iso) return ''
        const d = new Date(iso)
        return d.toLocaleDateString('es-MX', { month: 'long', day: 'numeric' })
    }

    const statusLabel =
        status === 'open' ? 'Abierto'
            : 'Cerrado'

    const balanceLabel = balance >= 0
        ? formatCurrency(balance)
        : `-${formatCurrency(Math.abs(balance))}`

    return (
        <div className="periodBar" role="status" aria-label="Resumen del periodo">
            <div className="periodBarLeft">
                <h2 className="periodBarTitle">Periodo actual</h2>
                <span className="periodBarDate">
                    {formatDate(startDate)} — {formatDate(endDate)}
                </span>
            </div>
            <div className="periodBarRight">
                <span className={`periodBarBalance${balance < 0 ? ' negative' : ''}`}>
                    {balanceLabel}
                </span>
                <span className={`periodBarStatus ${status}`}>
                    {statusLabel}
                </span>
            </div>
        </div>
    )
}

/**
 * PeriodStatusRibbon — legacy compact inline chip (kept for AppHeader backward compat).
 * @param {string} status - 'open' | 'review' | 'closed'
 */
const CHIP_LABELS = {
    open: 'Abierto',
    closed: 'Cerrado',
}

const CHIP_DESCRIPTIONS = {
    open: 'Periodo abierto — puedes registrar y editar gastos.',
    closed: 'Periodo cerrado — ya no se permiten cambios en gastos.',
}

export function PeriodStatusRibbon({ status = 'open' }) {
    const normalizedStatus = CHIP_LABELS[status] ? status : 'open'
    const label = CHIP_LABELS[normalizedStatus]
    const description = CHIP_DESCRIPTIONS[normalizedStatus]

    return (
        <span
            className={`periodChip periodChip-${normalizedStatus}`}
            role="status"
            aria-label={`Estado del periodo: ${description}`}
        >
            <span className="periodChipDot" aria-hidden>●</span>
            <span className="periodChipLabel">{label}</span>
        </span>
    )
}
