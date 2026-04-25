import { useMemo } from 'react'
import { formatCurrency } from '../utils'
import { EmptyState } from '../ui/EmptyState'
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
export function ExpenseList({
  items,
  isLoading,
  deletingId,
  savingId,
  onDelete,
  onSave,
  currency = 'MXN',
  isMutationLocked = false,
  onQuickAdd,
}) {
  const totalCents = useMemo(
    () => items.reduce((sum, item) => sum + item.amount_cents, 0),
    [items],
  )

  return (
    <section className="card expenseListCard" aria-label="Lista de gastos">
      {/* ── List header ── */}
      <div className="listHeader">
        <h2 className="listTitle">Gastos</h2>

        {!isLoading && items.length > 0 && (
          <div className="listMeta">
            <span className="listCount">
              {items.length} {items.length === 1 ? 'ítem' : 'ítems'}
            </span>
            <span className="totalBadge">{formatCurrency(totalCents, currency)}</span>
          </div>
        )}
      </div>

      {/* ── Loading skeletons ── */}
      {isLoading && (
        <ul className="expenseStack" aria-label="Cargando…" aria-busy>
          {[0, 1, 2].map((i) => (
            <li key={i} className="expenseItem skeleton" aria-hidden />
          ))}
        </ul>
      )}

      {/* ── Empty state ── */}
      {!isLoading && items.length === 0 && (
        <EmptyState
          title="Sin gastos aún"
          description="Agrega tu primer gasto usando añadir rápido."
          ctaLabel="Añadir rápido"
          onCta={onQuickAdd}
          icon="[+]"
          compact
        />
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
              isMutationLocked={isMutationLocked}
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
