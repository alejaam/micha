import { useMemo, useState } from 'react'
import { FormField } from '../ui/FormField'
import { Modal } from '../ui/Modal'
import { dollarsToCents } from '../utils'

/**
 * ExpenseModal — quick-add expense form inside a modal.
 * Shows only essential fields by default; advanced options in a collapsible section.
 *
 * Smart defaults: paymentMethod='card', expenseType='variable', isShared=true.
 * "Paid by" auto-defaults to the logged-in user's linked member.
 */
export function ExpenseModal({
    onClose,
    onSubmit,
    isSubmitting,
    members = [],
    isLoadingMembers = false,
    defaultPaidByMemberId = '',
}) {
    const [amount, setAmount] = useState('')
    const [description, setDescription] = useState('')
    const [isShared, setIsShared] = useState(true)
    const [paymentMethod, setPaymentMethod] = useState('card')
    const [expenseType, setExpenseType] = useState('variable')
    const [cardName, setCardName] = useState('')
    const [category, setCategory] = useState('other')
    const [showAdvanced, setShowAdvanced] = useState(false)

    const hasMembers = members.length > 0
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
        () => hasMembers && description.trim() !== '' && paidByMemberId.trim() !== '' && dollarsToCents(amount) !== null,
        [amount, description, paidByMemberId, hasMembers],
    )

    async function handleSubmit(e) {
        e.preventDefault()
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
    }

    return (
        <Modal title="New expense" onClose={onClose}>
            <form className="formStack" onSubmit={handleSubmit} noValidate>
                {/* Essential fields */}
                <FormField label="Amount" htmlFor="modalAmount">
                    <div className="inputWrap">
                        <span className="inputPrefix" aria-hidden>$</span>
                        <input
                            id="modalAmount"
                            className="input inputWithPrefix"
                            inputMode="decimal"
                            placeholder="0.00"
                            value={amount}
                            onChange={(e) => setAmount(e.target.value)}
                            autoFocus
                            disabled={isSubmitting}
                        />
                    </div>
                </FormField>

                <FormField label="Description" htmlFor="modalDescription">
                    <input
                        id="modalDescription"
                        className="input"
                        placeholder="e.g. Groceries at Trader Joe's"
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        autoComplete="off"
                        disabled={isSubmitting}
                    />
                </FormField>

                <FormField label="Category" htmlFor="modalCategory">
                    <select
                        id="modalCategory"
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

                {/* Advanced options toggle */}
                <button
                    type="button"
                    className="btn btnGhost btnSm advancedToggle"
                    onClick={() => setShowAdvanced((v) => !v)}
                >
                    {showAdvanced ? '▲ Hide options' : '▼ More options'}
                </button>

                {showAdvanced && (
                    <div className="formStack advancedSection">
                        {paidByMemberName ? <p className="formHint">Paid by: {paidByMemberName}</p> : null}

                        <FormField label="Payment method" htmlFor="modalPaymentMethod">
                            <select
                                id="modalPaymentMethod"
                                className="input"
                                value={paymentMethod}
                                onChange={(e) => setPaymentMethod(e.target.value)}
                                disabled={isSubmitting}
                            >
                                <option value="card">Card</option>
                                <option value="cash">Cash</option>
                                <option value="transfer">Transfer</option>
                                <option value="voucher">Voucher</option>
                            </select>
                        </FormField>

                        {isCardPayment && (
                            <FormField label="Card name" htmlFor="modalCardName">
                                <input
                                    id="modalCardName"
                                    className="input"
                                    placeholder="e.g. BANAMEX, HSBC, BBVA"
                                    value={cardName}
                                    onChange={(e) => setCardName(e.target.value)}
                                    autoComplete="off"
                                    disabled={isSubmitting}
                                />
                            </FormField>
                        )}

                        <FormField label="Expense type" htmlFor="modalExpenseType">
                            <select
                                id="modalExpenseType"
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

                        <label className="formLabel" style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                            <input
                                type="checkbox"
                                checked={isShared}
                                onChange={(e) => setIsShared(e.target.checked)}
                                disabled={isSubmitting}
                            />
                            Shared expense
                        </label>
                    </div>
                )}

                <div className="modalActions">
                    <button
                        type="submit"
                        className="btn btnPrimary btnFull"
                        disabled={!isValid || isSubmitting || isLoadingMembers}
                    >
                        {isSubmitting
                            ? <><span className="spinIcon" aria-hidden>⟳</span> Saving…</>
                            : 'Add expense'}
                    </button>
                    <button type="button" className="btn btnGhost btnFull" onClick={onClose} disabled={isSubmitting}>
                        Cancel
                    </button>
                </div>
            </form>
        </Modal>
    )
}
