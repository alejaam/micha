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
export function ExpenseForm({ onSubmit, isSubmitting, members = [], isLoadingMembers = false }) {
  const [amount, setAmount]                 = useState('')
  const [description, setDescription]       = useState('')
  const [paidByMemberId, setPaidByMemberId] = useState('')
  const [isShared, setIsShared]             = useState(true)
  const [paymentMethod, setPaymentMethod]   = useState('cash')
  const [expenseType, setExpenseType]       = useState('variable')

  const hasMembers = Array.isArray(members) && members.length > 0

  const isValid = useMemo(
    () => hasMembers && description.trim() !== '' && paidByMemberId.trim() !== '' && dollarsToCents(amount) !== null,
    [amount, description, paidByMemberId, hasMembers],
  )

  async function handleSubmit(event) {
    event.preventDefault()

    const amountCents = dollarsToCents(amount)
    if (amountCents === null) return

    await onSubmit({
      amountCents,
      description: description.trim(),
      paidByMemberId: paidByMemberId.trim(),
      isShared,
      paymentMethod,
      expenseType,
    })

    // Reset on successful submit (parent resolves the promise)
    setAmount('')
    setDescription('')
    setPaidByMemberId('')
    setIsShared(true)
    setPaymentMethod('cash')
    setExpenseType('variable')
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

        <FormField label="Paid by" htmlFor="newPaidByMemberId">
          <select
            id="newPaidByMemberId"
            className="input"
            value={paidByMemberId}
            onChange={(e) => setPaidByMemberId(e.target.value)}
            disabled={isSubmitting || isLoadingMembers || !hasMembers}
          >
            <option value="">
              {isLoadingMembers ? 'Loading members...' : hasMembers ? 'Select a member' : 'No members available'}
            </option>
            {members.map((member) => (
              <option key={member.id} value={member.id}>
                {member.name}
              </option>
            ))}
          </select>
        </FormField>

        {!isLoadingMembers && !hasMembers ? (
          <p className="formHint">Create at least one member before adding expenses.</p>
        ) : null}

        <FormField label="Payment method" htmlFor="newPaymentMethod">
          <select
            id="newPaymentMethod"
            className="input"
            value={paymentMethod}
            onChange={(e) => setPaymentMethod(e.target.value)}
            disabled={isSubmitting}
          >
            <option value="cash">cash</option>
            <option value="card">card</option>
            <option value="transfer">transfer</option>
            <option value="voucher">voucher</option>
          </select>
        </FormField>

        <FormField label="Expense type" htmlFor="newExpenseType">
          <select
            id="newExpenseType"
            className="input"
            value={expenseType}
            onChange={(e) => setExpenseType(e.target.value)}
            disabled={isSubmitting}
          >
            <option value="variable">variable</option>
            <option value="fixed">fixed</option>
            <option value="msi">msi</option>
          </select>
        </FormField>

        <label className="householdLabel" htmlFor="newIsShared" style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <input
            id="newIsShared"
            type="checkbox"
            checked={isShared}
            onChange={(e) => setIsShared(e.target.checked)}
            disabled={isSubmitting}
          />
          Shared expense
        </label>

        <button
          type="submit"
          className="btn btnPrimary btnFull"
          disabled={!isValid || isSubmitting || isLoadingMembers}
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
