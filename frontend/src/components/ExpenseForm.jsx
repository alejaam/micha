import { useMemo, useState } from 'react'
import { FormField } from '../ui/FormField'
import { dollarsToCents, sanitizeAmountInput } from '../utils'

/**
 * ExpenseForm — self-contained create-expense panel.
 * Manages its own draft state; calls onSubmit with validated {amountCents, description}.
 *
 * @param {(data:{amountCents:number,description:string})=>Promise<void>} onSubmit
 * @param {boolean} isSubmitting - Disables the form while a request is in-flight
 */
export function ExpenseForm({ onSubmit, isSubmitting, members = [], isLoadingMembers = false, defaultPaidByMemberId = '' }) {
  const [amount, setAmount]                 = useState('')
  const [description, setDescription]       = useState('')
  const [isShared, setIsShared]             = useState(true)
  const [paymentMethod, setPaymentMethod]   = useState('cash')
  const [expenseType, setExpenseType]       = useState('variable')
  const [cardName, setCardName]             = useState('')
  const [category, setCategory]             = useState('other')

  const hasMembers = Array.isArray(members) && members.length > 0
  const paidByMemberId = useMemo(
    () => defaultPaidByMemberId.trim() || members[0]?.id || '',
    [defaultPaidByMemberId, members],
  )
  const paidByMemberName = useMemo(
    () => members.find((member) => member.id === paidByMemberId)?.name || '',
    [members, paidByMemberId],
  )
  const isCardPayment = paymentMethod === 'card'

  const isValid = useMemo(
    () => hasMembers && description.trim() !== '' && paidByMemberId !== '' && dollarsToCents(amount) !== null,
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
      cardName: isCardPayment ? cardName.trim() : '',
      category,
    })

    // Reset on successful submit (parent resolves the promise)
    setAmount('')
    setDescription('')
    setIsShared(true)
    setPaymentMethod('cash')
    setExpenseType('variable')
    setCardName('')
    setCategory('other')
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
              onChange={(e) => setAmount(sanitizeAmountInput(e.target.value))}
              aria-label="Amount in dollars"
              disabled={isSubmitting}
              pattern="[0-9]*\.?[0-9]*"
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

        {!isLoadingMembers && paidByMemberName ? (
          <p className="formHint">Paid by: {paidByMemberName}</p>
        ) : null}

        {!isLoadingMembers && !hasMembers ? (
          <p className="formHint">Create at least one member before adding expenses.</p>
        ) : null}

        <FormField label="Category" htmlFor="newCategory">
          <select
            id="newCategory"
            className="input"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            disabled={isSubmitting}
          >
            <option value="rent">Rent</option>
            <option value="auto">Auto</option>
            <option value="streaming">Streaming / Services</option>
            <option value="food">Food</option>
            <option value="personal">Personal</option>
            <option value="savings">Savings</option>
            <option value="other">Other</option>
          </select>
        </FormField>

        <FormField label="Payment method" htmlFor="newPaymentMethod">
          <select
            id="newPaymentMethod"
            className="input"
            value={paymentMethod}
            onChange={(e) => setPaymentMethod(e.target.value)}
            disabled={isSubmitting}
          >
            <option value="cash">Cash</option>
            <option value="card">Card</option>
            <option value="transfer">Transfer</option>
            <option value="voucher">Voucher</option>
          </select>
        </FormField>

        {isCardPayment && (
          <FormField label="Card name" htmlFor="newCardName">
            <select
              id="newCardName"
              className="input"
              value={cardName}
              onChange={(e) => setCardName(e.target.value)}
              disabled={isSubmitting}
            >
              <option value="" disabled>Select card...</option>
              <option value="BBVA">BBVA</option>
              <option value="BANAMEX">Banamex</option>
              <option value="HSBC">HSBC</option>
              <option value="BANORTE">Banorte</option>
              <option value="SANTANDER">Santander</option>
              <option value="AMEX">Amex</option>
              <option value="NU">Nu</option>
              <option value="HEY BANCO">Hey Banco</option>
              <option value="RAPPI">Rappi</option>
              <option value="OTHER">Other</option>
            </select>
          </FormField>
        )}

        <FormField label="Expense type" htmlFor="newExpenseType">
          <select
            id="newExpenseType"
            className="input"
            value={expenseType}
            onChange={(e) => setExpenseType(e.target.value)}
            disabled={isSubmitting}
          >
            <option value="variable">Variable</option>
            <option value="fixed">Fixed</option>
            <option value="msi">MSI (installments)</option>
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
