import { useEffect, useMemo, useState } from 'react'
import { FormField } from '../ui/FormField'
import { Modal } from '../ui/Modal'
import { dollarsToCents, sanitizeAmountInput } from '../utils'
import { listCategories } from '../api'

/**
 * ExpenseModal — quick-add expense form inside a modal.
 *
 * Essential fields are always visible (amount, description, paid by, category, shared toggle).
 * Advanced options (payment method, expense type) are in a collapsible section with
 * contextual hints explaining the business impact of each choice.
 *
 * Categories are fetched dynamically from the backend per-household.
 */

const EXPENSE_TYPE_HINTS = {
    variable: 'One-time expense for this period only.',
    fixed: 'Recurs every period. Will be auto-copied when the current period closes.',
    msi: 'Installment purchase — will generate installments for the following months.',
}

export function ExpenseModal({
    onClose,
    onSubmit,
    isSubmitting,
    members = [],
    isLoadingMembers = false,
    defaultPaidByMemberId = '',
    householdId = '',
}) {
    const [amount, setAmount] = useState('')
    const [description, setDescription] = useState('')
    const [paidByMemberId, setPaidByMemberId] = useState(defaultPaidByMemberId.trim() || '')
    const [isShared, setIsShared] = useState(true)
    const [paymentMethod, setPaymentMethod] = useState('card')
    const [expenseType, setExpenseType] = useState('variable')
    const [totalInstallments, setTotalInstallments] = useState(3)
    const [cardName, setCardName] = useState('')
    const [category, setCategory] = useState('')
    const [showAdvanced, setShowAdvanced] = useState(false)

    // Dynamic categories from backend
    const [categories, setCategories] = useState([])
    const [loadingCategories, setLoadingCategories] = useState(false)

    useEffect(() => {
        if (!householdId) return
        let cancelled = false
        setLoadingCategories(true)
        listCategories({ householdId })
            .then((cats) => {
                if (!cancelled && Array.isArray(cats)) {
                    setCategories(cats)
                    // Default to 'other' slug or first category
                    const otherCat = cats.find((c) => c.slug === 'other')
                    if (!category) {
                        setCategory(otherCat?.id ?? otherCat?.slug ?? cats[0]?.id ?? cats[0]?.slug ?? 'other')
                    }
                }
            })
            .catch(() => {
                // Fallback: use hardcoded categories if backend fails
                if (!cancelled) {
                    setCategories([
                        { id: 'rent', slug: 'rent', name: 'Rent', is_default: true },
                        { id: 'auto', slug: 'auto', name: 'Auto', is_default: true },
                        { id: 'streaming', slug: 'streaming', name: 'Streaming / Services', is_default: true },
                        { id: 'food', slug: 'food', name: 'Food', is_default: true },
                        { id: 'personal', slug: 'personal', name: 'Personal', is_default: true },
                        { id: 'savings', slug: 'savings', name: 'Savings', is_default: true },
                        { id: 'other', slug: 'other', name: 'Other', is_default: true },
                    ])
                    if (!category) setCategory('other')
                }
            })
            .finally(() => {
                if (!cancelled) setLoadingCategories(false)
            })
        return () => { cancelled = true }
    }, [householdId])

    const hasMembers = members.length > 0
    const isCardPayment = paymentMethod === 'card'
    const isVoucher = paymentMethod === 'voucher'
    const isMSI = expenseType === 'msi'

    // Sync paidByMemberId when members load or defaultPaidByMemberId changes
    useEffect(() => {
        if (members.length > 0 && !paidByMemberId) {
            setPaidByMemberId(defaultPaidByMemberId || members[0].id)
        }
    }, [members, paidByMemberId, defaultPaidByMemberId])

    const isCurrentMemberSelected = paidByMemberId === defaultPaidByMemberId && defaultPaidByMemberId !== ''

    const isValid = useMemo(
        () => {
            const basic = hasMembers && description.trim() !== '' && paidByMemberId.trim() !== '' && dollarsToCents(amount) !== null
            if (!basic) return false
            if (isMSI && (isNaN(totalInstallments) || totalInstallments <= 0)) return false
            return true
        },
        [amount, description, paidByMemberId, hasMembers, isMSI, totalInstallments],
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
            totalInstallments: isMSI ? Number(totalInstallments) : 0,
        })
    }

    return (
        <Modal title="New expense" onClose={onClose}>
            <form className="formStack" onSubmit={handleSubmit} noValidate>
                {/* ── Essential fields ── */}
                <FormField label="Amount" htmlFor="modalAmount">
                    <div className="inputWrap">
                        <span className="inputPrefix" aria-hidden>$</span>
                        <input
                            id="modalAmount"
                            className="input inputWithPrefix"
                            inputMode="decimal"
                            placeholder="0.00"
                            value={amount}
                            onChange={(e) => setAmount(sanitizeAmountInput(e.target.value))}
                            autoFocus
                            disabled={isSubmitting}
                            pattern="[0-9]*\.?[0-9]*"
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

                {hasMembers && (
                    <p className="formHint">
                        Paid by: <strong>{members.find((m) => m.id === paidByMemberId)?.name || 'You'}</strong> (you)
                    </p>
                )}

                <FormField label="Category" htmlFor="modalCategory">
                    <select
                        id="modalCategory"
                        className="input"
                        value={category}
                        onChange={(e) => setCategory(e.target.value)}
                        disabled={isSubmitting || loadingCategories}
                    >
                        {loadingCategories ? (
                            <option>Loading…</option>
                        ) : (
                            categories.map((c) => (
                                <option key={c.id || c.slug} value={c.id || c.slug}>
                                    {c.name}
                                </option>
                            ))
                        )}
                    </select>
                </FormField>

                {/* ── Shared toggle — promoted from advanced ── */}
                <div className="sharedToggle">
                    <label className="sharedToggleLabel" htmlFor="modalShared">
                        <input
                            id="modalShared"
                            type="checkbox"
                            checked={isShared}
                            onChange={(e) => setIsShared(e.target.checked)}
                            disabled={isSubmitting}
                        />
                        <span className="sharedToggleText">Shared expense</span>
                    </label>
                    <p className="formHint">
                        {isShared
                            ? 'Will be split among household members according to the settlement mode.'
                            : 'Personal expense — will NOT be included in settlement calculations.'}
                    </p>
                </div>

                {/* ── Advanced options toggle ── */}
                <button
                    type="button"
                    className="btn btnGhost btnSm advancedToggle"
                    onClick={() => setShowAdvanced((v) => !v)}
                >
                    {showAdvanced ? '▲ Hide options' : '▼ More options'}
                </button>

                {showAdvanced && (
                    <div className="formStack advancedSection">
                        <FormField label="Payment method" htmlFor="modalPaymentMethod">
                            <select
                                id="modalPaymentMethod"
                                className="input"
                                value={paymentMethod}
                                onChange={(e) => setPaymentMethod(e.target.value)}
                                disabled={isSubmitting}
                            >
                                <option value="card">💳 Card</option>
                                <option value="cash">💵 Cash</option>
                                <option value="transfer">🏦 Transfer</option>
                                <option value="voucher">🎟️ Voucher</option>
                            </select>
                            {isVoucher && (
                                <p className="formHint">
                                    Voucher expenses are included in settlement calculations.
                                </p>
                            )}
                        </FormField>

                        {isCardPayment && (
                            <FormField label="Card name" htmlFor="modalCardName">
                                <select
                                    id="modalCardName"
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

                        <FormField label="Expense type" htmlFor="modalExpenseType">
                            <select
                                id="modalExpenseType"
                                className="input"
                                value={expenseType}
                                onChange={(e) => setExpenseType(e.target.value)}
                                disabled={isSubmitting}
                            >
                                <option value="variable">📝 Variable</option>
                                <option value="fixed">📌 Fixed (recurrent)</option>
                                <option value="msi">🔒 MSI (installments)</option>
                            </select>
                            <p className="formHint">{EXPENSE_TYPE_HINTS[expenseType]}</p>
                        </FormField>

                        {isMSI && (
                            <FormField label="Total installments" htmlFor="modalTotalInstallments">
                                <input
                                    id="modalTotalInstallments"
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
