import { formatCurrency, formatRelativeDate } from '../utils'
import { getCategoryIcon } from '../utils/categoryIcons'
import { EmptyState } from '../ui/EmptyState'

/**
 * RecentExpenses — shows the last N expense items in a compact read-only list.
 *
 * @param {Array}   items    - Expense objects (already sorted desc by date)
 * @param {boolean} isLoading
 * @param {string}  currency
 * @param {number}  [limit=8] - How many to show
 */
export function RecentExpenses({ items, isLoading, currency = 'MXN', limit = 8, onQuickAdd }) {
    if (isLoading) {
        return (
            <ul className="expenseStack" aria-label="Loading recent expenses" aria-busy>
                {[0, 1, 2].map((i) => (
                    <li key={i} className="expenseItem skeleton" aria-hidden />
                ))}
            </ul>
        )
    }

    const visible = items.slice(0, limit)

    if (visible.length === 0) {
        return (
            <EmptyState
                title="Sin gastos recientes"
                description="Agrega tu primer gasto para ver tendencias e historial."
                ctaLabel="Añadir gasto"
                onCta={onQuickAdd}
                icon="[+]"
                compact
            />
        )
    }

    return (
        <ul className="expenseStack">
            {visible.map((item, index) => {
                const delayMs = Math.min(index * 35, 250)
                const categoryKey = item.category || item.category_slug || item.category_name || 'other'
                return (
                    <li
                        key={item.id}
                        className="expenseItem slideUp"
                        style={{ animationDelay: `${delayMs}ms` }}
                    >
                        <div className="expenseBody">
                            <span className="expenseDesc">{item.description}</span>
                            <span className="expenseMeta">
                                {getCategoryIcon(categoryKey)} · 
                                {formatRelativeDate(item.created_at)}
                                {item.expense_type ? ` · ${item.expense_type}` : ''}
                            </span>
                        </div>
                        <div className="expenseRight">
                            <span className="expenseAmount">
                                {formatCurrency(item.amount_cents, item.currency || currency)}
                            </span>
                        </div>
                    </li>
                )
            })}
        </ul>
    )
}
