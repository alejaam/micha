import { formatCurrency, formatRelativeDate } from '../utils'
import { EmptyState } from '../ui/EmptyState'

/**
 * Feed — redesigned RecentExpenses. Renders a chronological activity feed
 * with member avatars, type icons, descriptions, and amounts.
 *
 * @param {Array}   items      - Expense objects (sorted desc by date)
 * @param {boolean} isLoading
 * @param {string}  currency
 * @param {Array}   members    - Member list (for avatar/name lookup)
 * @param {number}  [limit=10] - How many to show
 * @param {Function} onQuickAdd
 */
export function Feed({ items = [], isLoading = false, currency = 'MXN', members = [], limit = 10, onQuickAdd }) {
    const memberMap = Object.fromEntries(
        Array.isArray(members) ? members.map((m) => [m.id, m.name]) : [],
    )

    if (isLoading) {
        return (
            <ul className="feed" aria-label="Cargando movimientos" aria-busy>
                {[0, 1, 2].map((i) => (
                    <li key={i} className="feedItem skeleton" aria-hidden />
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

    const getTypeIcon = (type) => {
        switch (type) {
            case 'fixed':
                return '\uD83D\uDD04'   // 🔄
            case 'msi':
                return '\uD83D\uDCC6'   // 📆
            case 'variable':
            default:
                return '\u270F\uFE0F'   // ✏️
        }
    }

    return (
        <ul className="feed">
            {visible.map((item) => {
                const memberName = memberMap[item.paid_by_member_id] || 'Alguien'
                const initial = memberName.charAt(0).toUpperCase()
                const typeIcon = getTypeIcon(item.expense_type)

                return (
                    <li key={item.id} className="feedItem">
                        <div className="feedItemAvatar">{initial}</div>
                        <div className="feedItemBody">
                            <span className="feedItemName">{memberName}</span>
                            <span className="feedItemDesc">
                                pagó {formatCurrency(item.amount_cents, item.currency || currency)} · {item.description}
                            </span>
                            <span className="feedItemMeta">
                                <span className="feedItemType">{typeIcon}</span>
                                <span className="feedTime">{formatRelativeDate(item.created_at)}</span>
                            </span>
                        </div>
                        <span className="feedItemAmount">
                            {formatCurrency(item.amount_cents, item.currency || currency)}
                        </span>
                    </li>
                )
            })}
        </ul>
    )
}
