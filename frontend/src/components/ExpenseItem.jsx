import { useMemo, useState } from 'react'
import { FormField } from '../ui/FormField'
import { dollarsToCents, formatCurrency, formatRelativeDate } from '../utils'

/**
 * ExpenseItem — a single row in the expense list.
 * Toggles between read mode (amount + actions) and inline edit mode.
 *
 * @param {{ id:string, description:string, amount_cents:number, created_at?:string }} item
 * @param {boolean} isDeleting   - Shows spinner on the delete button
 * @param {boolean} isSaving     - Shows spinner on the save button
 * @param {(id:string)=>void}     onDelete
 * @param {({id,amountCents,description})=>void} onSave
 * @param {number} animIndex     - Staggered entrance delay index
 */
export function ExpenseItem({ item, isDeleting, isSaving, onDelete, onSave, animIndex, currency = 'MXN' }) {
  const [editing, setEditing]         = useState(false)
  const [draftAmount, setDraftAmount] = useState('')
  const [draftDesc, setDraftDesc]     = useState('')

  const isDraftValid = useMemo(
    () => draftDesc.trim() !== '' && dollarsToCents(draftAmount) !== null,
    [draftAmount, draftDesc],
  )

  function startEdit() {
    setDraftAmount((item.amount_cents / 100).toFixed(2))
    setDraftDesc(item.description)
    setEditing(true)
  }

  function cancelEdit() {
    setEditing(false)
  }

  async function handleSave() {
    const amountCents = dollarsToCents(draftAmount)
    if (amountCents === null) return
    await onSave({ id: item.id, amountCents, description: draftDesc.trim() })
    setEditing(false)
  }

  const delayMs = Math.min(animIndex * 45, 300)

  return (
    <li
      className="expenseItem slideUp"
      style={{ animationDelay: `${delayMs}ms` }}
    >
      {/* Read mode */}
      {!editing && (
        <>
          <div className="expenseBody">
            <span className="expenseDesc">{item.description}</span>
            <span className="expenseMeta">
              {formatRelativeDate(item.created_at)}
              {item.created_at && ' · '}
              {item.expense_type ? `${item.expense_type} · ` : ''}
              {item.total_installments > 0 ? `${item.total_installments} inst. · ` : ''}
              <span className="expenseId">{item.id.slice(0, 8)}…</span>
            </span>
          </div>

          <div className="expenseRight">
            <span className="expenseAmount">{formatCurrency(item.amount_cents, item.currency || currency)}</span>
            <div className="expenseActions">
              <button
                type="button"
                className="btn btnGhost btnSm btnIcon"
                onClick={startEdit}
                disabled={isDeleting}
                aria-label={`Edit ${item.description}`}
                title="Edit"
              >
                ✎
              </button>
              <button
                type="button"
                className="btn btnGhostDanger btnSm btnIcon"
                onClick={() => onDelete(item.id)}
                disabled={isDeleting}
                aria-label={`Delete ${item.description}`}
                title="Delete"
              >
                {isDeleting ? <span className="spinIcon">⟳</span> : '✕'}
              </button>
            </div>
          </div>
        </>
      )}

      {/* Inline edit mode */}
      {editing && (
        <div className="editBox">
          <div className="editFields">
            <FormField label="Amount" htmlFor={`editAmt-${item.id}`}>
              <div className="inputWrap">
                <span className="inputPrefix" aria-hidden>$</span>
                <input
                  id={`editAmt-${item.id}`}
                  className="input inputWithPrefix inputSm"
                  inputMode="decimal"
                  value={draftAmount}
                  onChange={(e) => setDraftAmount(e.target.value)}
                  aria-label="Edit amount"
                  disabled={isSaving}
                />
              </div>
            </FormField>

            <FormField label="Description" htmlFor={`editDesc-${item.id}`}>
              <input
                id={`editDesc-${item.id}`}
                className="input inputSm"
                value={draftDesc}
                onChange={(e) => setDraftDesc(e.target.value)}
                aria-label="Edit description"
                disabled={isSaving}
              />
            </FormField>
          </div>

          <div className="editActions">
            <button
              type="button"
              className="btn btnPrimary btnSm"
              onClick={handleSave}
              disabled={!isDraftValid || isSaving}
            >
              {isSaving ? <><span className="spinIcon">⟳</span> Saving…</> : 'Save'}
            </button>
            <button
              type="button"
              className="btn btnGhost btnSm"
              onClick={cancelEdit}
              disabled={isSaving}
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </li>
  )
}
