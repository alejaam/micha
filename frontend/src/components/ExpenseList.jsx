import { useMemo } from 'react'
import { formatCurrency } from '../utils'
import { ExpenseItem } from './ExpenseItem'

/**
 * ExpenseList — card that holds the list header, skeleton loaders,
 * empty state, and the scrollable expense rows.
 *
 * @param {Array}   items        - Loaded expense objects from the API
 * @param {boolean} isLoading    - Show skeleton placeholders
 * @param {string}  deletingId   - ID of item currently being deleted
 * @param {string}  savingId     - ID of item currently being saved
 * @param {(id:string)=>void}     onDelete
 * @param {(data)=>void}          onSave
 */
export function ExpenseList({ items, isLoading, deletingId, savingId, onDelete, onSave, currency = 'MXN' }) {
  const totalCents = useMemo(
    () => items.reduce((sum, item) => sum + item.amount_cents, 0),
    [items],
  )

  return (
    <section className="card expenseListCard" aria-label="Expense list">
      {/* ── List header ── */}
      <div className="listHeader">
        <h2 className="listTitle">Expenses</h2>

        {!isLoading && items.length > 0 && (
          <div className="listMeta">
            <span className="listCount">
              {items.length} item{items.length !== 1 ? 's' : ''}
            </span>
            <span className="totalBadge">{formatCurrency(totalCents, currency)}</span>
          </div>
        )}
      </div>

      {/* ── Loading skeletons ── */}
      {isLoading && (
        <ul className="expenseStack" aria-label="Loading…" aria-busy>
          {[0, 1, 2].map((i) => (
            <li key={i} className="expenseItem skeleton" aria-hidden />
          ))}
        </ul>
      )}

      {/* ── Empty state ── */}
      {!isLoading && items.length === 0 && (
        <div className="emptyState">
          <div className="emptyIcon" aria-hidden>[]</div>
          <p className="emptyTitle">No expenses yet</p>
          <p className="emptyHint">Add your first expense using the form.</p>
        </div>
      )}

      {/* ── Expense rows ── */}
      {!isLoading && items.length > 0 && (
        <ul className="expenseStack">
          {items.map((item, index) => (
            <ExpenseItem
              key={item.id}
              item={item}
              animIndex={index}
              isDeleting={deletingId === item.id}
              isSaving={savingId === item.id}
              onDelete={onDelete}
              onSave={onSave}
              currency={currency}
            />
          ))}
        </ul>
      )}
    </section>
  )
}
