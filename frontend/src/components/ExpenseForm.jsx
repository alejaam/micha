import { useMemo, useState } from 'react'
import { FormField } from '../ui/FormField'
import { dollarsToCents } from '../utils'

/**
 * ExpenseForm — self-contained create-expense panel.
 * Manages its own draft state; calls onSubmit with validated {amountCents, description}.
 *
 * @param {(data:{amountCents:number,description:string})=>Promise<void>} onSubmit
 * @param {boolean} isSubmitting - Disables the form while a request is in-flight
 */
export function ExpenseForm({ onSubmit, isSubmitting }) {
  const [amount, setAmount]           = useState('')
  const [description, setDescription] = useState('')

  const isValid = useMemo(
    () => description.trim() !== '' && dollarsToCents(amount) !== null,
    [amount, description],
  )

  async function handleSubmit(event) {
    event.preventDefault()

    const amountCents = dollarsToCents(amount)
    if (amountCents === null) return

    await onSubmit({ amountCents, description: description.trim() })

    // Reset on successful submit (parent resolves the promise)
    setAmount('')
    setDescription('')
  }

  return (
    <section className="card" aria-label="Add a new expense">
      <h2 className="sectionTitle">
        <span className="sectionTitleIcon" aria-hidden>＋</span>
        New expense
      </h2>

      <form onSubmit={handleSubmit} className="formStack" noValidate>
        <FormField label="Amount" htmlFor="newAmount">
          <div className="inputWrap">
            <span className="inputPrefix" aria-hidden>$</span>
            <input
              id="newAmount"
              className="input inputWithPrefix"
              inputMode="decimal"
              placeholder="0.00"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              aria-label="Amount in dollars"
              disabled={isSubmitting}
            />
          </div>
        </FormField>

        <FormField label="Description" htmlFor="newDescription">
          <input
            id="newDescription"
            className="input"
            placeholder="e.g. Groceries at Trader Joe's"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            autoComplete="off"
            disabled={isSubmitting}
          />
        </FormField>

        <button
          type="submit"
          className="btn btnPrimary btnFull"
          disabled={!isValid || isSubmitting}
        >
          {isSubmitting ? (
            <>
              <span className="spinIcon" aria-hidden>⟳</span>
              Saving…
            </>
          ) : (
            'Add expense'
          )}
        </button>
      </form>
    </section>
  )
}
