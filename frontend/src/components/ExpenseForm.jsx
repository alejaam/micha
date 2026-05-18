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
export function ExpenseForm({ onSubmit, isSubmitting, members = [], cards = [], isLoadingMembers = false, defaultPaidByMemberId = '' }) {
  const [amount, setAmount]                 = useState('')
  const [description, setDescription]       = useState('')
  const [isShared, setIsShared]             = useState(true)
  const [paymentMethod, setPaymentMethod]   = useState('cash')

  const [expenseType, setExpenseType]       = useState('variable')
  const [totalInstallments, setTotalInstallments] = useState(3)
  const [cardId, setCardId]                 = useState('')
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
  const isMSI = expenseType === 'msi'

  const isValid = useMemo(
    () => {
      const basic = hasMembers && description.trim() !== '' && paidByMemberId !== '' && dollarsToCents(amount) !== null
      if (!basic) return false
      if (isMSI && (isNaN(totalInstallments) || totalInstallments <= 0)) return false
      if (isCardPayment && cards.length === 0) return false
      if (isCardPayment && cardId === '') return false
      return true
    },
    [amount, description, paidByMemberId, hasMembers, isMSI, totalInstallments, isCardPayment, cards.length, cardId],
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
      cardId: isCardPayment ? cardId : '',
      category,
      totalInstallments: isMSI ? Number(totalInstallments) : 0,
    })

    // Reset on successful submit (parent resolves the promise)
    setAmount('')
    setDescription('')
    setIsShared(true)
    setPaymentMethod('cash')
    setExpenseType('variable')
    setTotalInstallments(3)
    setCardId('')
    setCategory('other')
  }

  return (
    <section className="card" aria-label="Agregar nuevo gasto">
      <h2 className="sectionTitle">
        <span className="sectionTitleIcon" aria-hidden>＋</span>
        Nuevo gasto
      </h2>

      <form onSubmit={handleSubmit} className="formStack" noValidate>
        <FormField label="Monto" htmlFor="newAmount">
          <div className="inputWrap">
            <span className="inputPrefix" aria-hidden>$</span>
            <input
              id="newAmount"
              className="input inputWithPrefix"
              inputMode="decimal"
              placeholder="0.00"
              value={amount}
              onChange={(e) => setAmount(sanitizeAmountInput(e.target.value))}
              aria-label="Monto"
              disabled={isSubmitting}
              pattern="[0-9]*\.?[0-9]*"
            />
          </div>
        </FormField>

        <FormField label="Descripción" htmlFor="newDescription">
          <input
            id="newDescription"
            className="input"
            placeholder="Ej. Súper de la semana"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            autoComplete="off"
            disabled={isSubmitting}
          />
        </FormField>

        {!isLoadingMembers && paidByMemberName ? (
          <p className="formHint">Pagado por: {paidByMemberName}</p>
        ) : null}

        {!isLoadingMembers && !hasMembers ? (
          <p className="formHint">Crea al menos un miembro antes de agregar gastos.</p>
        ) : null}

        <FormField label="Categoría" htmlFor="newCategory">
          <select
            id="newCategory"
            className="input"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            disabled={isSubmitting}
          >
            <option value="rent">Renta</option>
            <option value="auto">Auto</option>
            <option value="streaming">Streaming / Servicios</option>
            <option value="food">Comida</option>
            <option value="personal">Personal</option>
            <option value="savings">Ahorros</option>
            <option value="other">Otro</option>
          </select>
        </FormField>

        <FormField label="Método de pago" htmlFor="newPaymentMethod">
          <select
            id="newPaymentMethod"
            className="input"
            value={paymentMethod}
            onChange={(e) => setPaymentMethod(e.target.value)}
            disabled={isSubmitting}
          >
            <option value="cash">Efectivo</option>
            <option value="card">Tarjeta</option>
            <option value="transfer">Transferencia</option>
            <option value="voucher">Vale</option>
          </select>
        </FormField>

        {isCardPayment && (
          <FormField label="Tarjeta" htmlFor="newCardId">
            {cards.length > 0 ? (
              <select
                id="newCardId"
                className="input"
                value={cardId}
                onChange={(e) => setCardId(e.target.value)}
                disabled={isSubmitting}
              >
                <option value="" disabled>Selecciona tarjeta...</option>
                {cards.map((card) => (
                  <option key={card.id} value={card.id}>{card.card_name}</option>
                ))}
              </select>
            ) : (
              <p className="formHint formHintError">No tienes tarjetas registradas. Agrégalas en la sección de administración.</p>
            )}
          </FormField>
        )}

        <FormField label="Tipo de gasto" htmlFor="newExpenseType">
          <select
            id="newExpenseType"
            className="input"
            value={expenseType}
            onChange={(e) => setExpenseType(e.target.value)}
            disabled={isSubmitting}
          >
            <option value="variable">Variable</option>
            <option value="fixed">Fijo</option>
            <option value="msi">MSI (meses sin intereses)</option>
          </select>
        </FormField>

        {isMSI && (
          <FormField label="Total de cuotas" htmlFor="newTotalInstallments">
            <input
              id="newTotalInstallments"
              className="input"
              type="number"
              min="1"
              max="48"
              value={totalInstallments}
              onChange={(e) => setTotalInstallments(e.target.value)}
              disabled={isSubmitting}
            />
          </FormField>
        )}

        <label className="householdLabel" htmlFor="newIsShared" style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <input
            id="newIsShared"
            type="checkbox"
            checked={isShared}
            onChange={(e) => setIsShared(e.target.checked)}
            disabled={isSubmitting}
          />
          Gasto compartido
        </label>

        <button
          type="submit"
          className="btn btnPrimary btnFull"
          disabled={!isValid || isSubmitting || isLoadingMembers}
        >
          {isSubmitting ? (
            <>
              <span className="spinIcon" aria-hidden>⟳</span>
              Guardando…
            </>
          ) : (
            'Agregar gasto'
          )}
        </button>
      </form>
    </section>
  )
}
