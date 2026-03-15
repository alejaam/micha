import { formatCurrency, formatRelativeDate } from '../utils'

/**
 * RecentExpenses — shows the last N expense items in a compact read-only list.
 *
 * @param {Array}   items    - Expense objects (already sorted desc by date)
 * @param {boolean} isLoading
 * @param {string}  currency
 * @param {number}  [limit=8] - How many to show
 */
export function RecentExpenses({ items, isLoading, currency = 'MXN', limit = 8 }) {
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
            <div className="emptyState">
                <div className="emptyIcon" aria-hidden>🧾</div>
                <p className="emptyTitle">No expenses yet</p>
                <p className="emptyHint">Tap + to add your first expense.</p>
            </div>
        )
    }

    return (
        <ul className="expenseStack">
            {visible.map((item, index) => {
                const delayMs = Math.min(index * 35, 250)
                return (
                    <li
                        key={item.id}
                        className="expenseItem slideUp"
                        style={{ animationDelay: `${delayMs}ms` }}
                    >
                        <div className="expenseBody">
                            <span className="expenseDesc">{item.description}</span>
                            <span className="expenseMeta">
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
