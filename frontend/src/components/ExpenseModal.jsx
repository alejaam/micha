import { useMemo, useState } from 'react'
import { Modal } from '../ui/Modal'
import { FormField } from '../ui/FormField'
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
    const [paidByMemberId, setPaidByMemberId] = useState(defaultPaidByMemberId)
    const [isShared, setIsShared] = useState(true)
    const [paymentMethod, setPaymentMethod] = useState('card')
    const [expenseType, setExpenseType] = useState('variable')
    const [showAdvanced, setShowAdvanced] = useState(false)

    const hasMembers = members.length > 0

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
                        <FormField label="Paid by" htmlFor="modalPaidBy">
                            <select
                                id="modalPaidBy"
                                className="input"
                                value={paidByMemberId}
                                onChange={(e) => setPaidByMemberId(e.target.value)}
                                disabled={isSubmitting || isLoadingMembers || !hasMembers}
                            >
                                <option value="">
                                    {isLoadingMembers ? 'Loading…' : hasMembers ? 'Select a member' : 'No members'}
                                </option>
                                {members.map((m) => (
                                    <option key={m.id} value={m.id}>{m.name}</option>
                                ))}
                            </select>
                        </FormField>

                        <FormField label="Payment method" htmlFor="modalPaymentMethod">
                            <select
                                id="modalPaymentMethod"
                                className="input"
                                value={paymentMethod}
                                onChange={(e) => setPaymentMethod(e.target.value)}
                                disabled={isSubmitting}
                            >
                                <option value="card">card</option>
                                <option value="cash">cash</option>
                                <option value="transfer">transfer</option>
                                <option value="voucher">voucher</option>
                            </select>
                        </FormField>

                        <FormField label="Expense type" htmlFor="modalExpenseType">
                            <select
                                id="modalExpenseType"
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
