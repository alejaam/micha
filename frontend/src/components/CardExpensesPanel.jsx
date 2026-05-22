import { useMemo } from 'react'
import { formatCurrency } from '../utils'

/**
 * MSI — redesigned CardExpensesPanel. Renders a compact list of
 * monthly installment payments with progress bars and fraction labels.
 *
 * @param {Array}  items       - Raw expense list (for amount lookup)
 * @param {Array}  msiProgress - Derived MSI progress data from useDashboardDerivedData
 * @param {string} currency
 */
export function MSI({ items = [], msiProgress = [], currency = 'MXN' }) {
    // Build a lookup of MSI expense amounts to compute "X/mes"
    const msiAmountMap = useMemo(() => {
        const map = {}
        if (!Array.isArray(items)) return map
        for (const item of items) {
            if (item.expense_type === 'msi' && item.total_installments > 0) {
                map[item.id] = Math.round(item.amount_cents / Number(item.total_installments))
            }
        }
        return map
    }, [items])

    const list = Array.isArray(msiProgress) ? msiProgress : []

    if (list.length === 0) {
        return (
            <section className="card" aria-label="Pagos a meses">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>📆</span>
                    Pagos a meses (MSI)
                </h2>
                <div className="emptyState">
                    <p className="emptyTitle">Sin MSI activos</p>
                    <p className="emptyHint">
                        Los gastos a meses sin intereses aparecerán aquí.
                    </p>
                </div>
            </section>
        )
    }

    return (
        <section className="card" aria-label="Pagos a meses">
            <h2 className="sectionTitle">
                <span className="sectionTitleIcon" aria-hidden>📆</span>
                Pagos a meses (MSI)
                <span className="sectionBadge">{list.length} activos</span>
            </h2>
            <div className="msiList">
                {list.slice(0, 10).map((msi) => {
                    const monthlyCents = msiAmountMap[msi.id] || 0
                    return (
                        <div key={msi.id} className="msiItem">
                            <div className="msiItemHeader">
                                <span className="msiItemName">{msi.description}</span>
                                <span className="msiItemAmount">
                                    {formatCurrency(monthlyCents, currency)}/mes
                                </span>
                            </div>
                            <div className="msiItemProgress">
                                <div className="msiBar" aria-hidden>
                                    <div
                                        className="msiBarFill"
                                        style={{ width: `${Math.min(msi.progressPercent, 100)}%` }}
                                    />
                                </div>
                                <span className="msiLabel">
                                    {msi.currentInstallment}/{msi.totalInstallments}
                                </span>
                            </div>
                        </div>
                    )
                })}
            </div>
        </section>
    )
}
