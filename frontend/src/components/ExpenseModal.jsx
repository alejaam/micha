import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { listCards, listCategories } from '../api'
import { MEXICAN_BANKS } from '../constants/mexicanBanks'
import { FormField } from '../ui/FormField'
import { Modal } from '../ui/Modal'
import { Tooltip } from '../ui/Tooltip'
import { dollarsToCents, sanitizeAmountInput } from '../utils'
import { getCategoryIcon } from '../utils/categoryIcons'

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
    variable: 'Gasto único para este periodo.',
    msi: 'Compra a meses sin intereses; se generan cuotas para los meses siguientes.',
}

function preferredCardStorageKey(householdId) {
    return `micha_preferred_card_${householdId}`
}

export function ExpenseModal({
    onClose,
    onSubmit,
    isSubmitting,
    isMutationLocked = false,
    members = [],
    isLoadingMembers = false,
    defaultPaidByMemberId = '',
    householdId = '',
}) {
    const navigate = useNavigate()
    const [amount, setAmount] = useState('')
    const [description, setDescription] = useState('')
    const [paidByMemberId, setPaidByMemberId] = useState(defaultPaidByMemberId.trim() || '')
    const [isShared, setIsShared] = useState(true)
    const [paymentMethod, setPaymentMethod] = useState('card')
    const [expenseType, setExpenseType] = useState('variable')
    const [totalInstallments, setTotalInstallments] = useState(3)
    const [cardId, setCardId] = useState('')
    const [cardName, setCardName] = useState('')
    const [category, setCategory] = useState('')
    const [showAdvanced, setShowAdvanced] = useState(false)

    // Dynamic categories from backend
    const [categories, setCategories] = useState([])
    const [loadingCategories, setLoadingCategories] = useState(false)
    const [cards, setCards] = useState([])
    const [loadingCards, setLoadingCards] = useState(false)

    const isCurrentUserOwner = useMemo(() => {
        if (!defaultPaidByMemberId) return false

        const linkedMembers = members
            .filter((member) => member.user_id && String(member.user_id).trim() !== '')
            .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())

        if (linkedMembers.length === 0) return false

        return linkedMembers[0].id === defaultPaidByMemberId
    }, [members, defaultPaidByMemberId])

    const eligibleMembers = useMemo(() => {
        if (isCurrentUserOwner) {
            return members
        }

        return members.filter((member) => member.id === defaultPaidByMemberId)
    }, [members, isCurrentUserOwner, defaultPaidByMemberId])

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
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [householdId])

    useEffect(() => {
        if (!householdId) return
        let cancelled = false
        setLoadingCards(true)
        listCards({ householdId })
            .then((items) => {
                if (cancelled || !Array.isArray(items)) return
                setCards(items)
                if (items.length > 0 && !cardId) {
                    const preferredCardId = localStorage.getItem(preferredCardStorageKey(householdId)) ?? ''
                    const preferredCard = items.find((item) => item.id === preferredCardId)
                    const initialCard = preferredCard ?? items[0]
                    setCardId(initialCard.id)
                    setCardName(initialCard.card_name || '')
                }
            })
            .catch(() => {
                if (!cancelled) {
                    setCards([])
                }
            })
            .finally(() => {
                if (!cancelled) setLoadingCards(false)
            })
        return () => {
            cancelled = true
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [householdId])

    useEffect(() => {
        if (!householdId || !cardId) return
        localStorage.setItem(preferredCardStorageKey(householdId), cardId)
    }, [householdId, cardId])

    const hasMembers = eligibleMembers.length > 0
    const isCardPayment = paymentMethod === 'card'
    const isVoucher = paymentMethod === 'voucher'
    const isMSI = expenseType === 'msi'
    const hasRegisteredCards = cards.length > 0

    useEffect(() => {
        if (paymentMethod === 'card') return
        setCardId('')
        setCardName('')
    }, [paymentMethod])

    // Sync paidByMemberId when members load or defaultPaidByMemberId changes
    useEffect(() => {
        if (eligibleMembers.length > 0 && !paidByMemberId) {
            const preferredMember = eligibleMembers.find((member) => member.id === defaultPaidByMemberId)
            setPaidByMemberId(preferredMember?.id || eligibleMembers[0].id)
        }
    }, [eligibleMembers, paidByMemberId, defaultPaidByMemberId])

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
        if (isMutationLocked) return
        const amountCents = dollarsToCents(amount)
        if (amountCents === null) return

        await onSubmit({
            amountCents,
            description: description.trim(),
            paidByMemberId: paidByMemberId.trim(),
            isShared,
            paymentMethod,
            expenseType,
            cardId: isCardPayment ? cardId.trim() : '',
            cardName: isCardPayment ? cardName.trim() : '',
            category,
            totalInstallments: isMSI ? Number(totalInstallments) : 0,
        })
    }

    return (
        <Modal title="Nuevo gasto" onClose={onClose}>
            <form className="formStack" onSubmit={handleSubmit} noValidate>
                {/* ── Essential fields ── */}
                <FormField label="Monto" htmlFor="modalAmount">
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

                <FormField label="Descripción" htmlFor="modalDescription">
                    <input
                        id="modalDescription"
                        className="input"
                        placeholder="Ej. Súper de la semana"
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        autoComplete="off"
                        disabled={isSubmitting}
                    />
                </FormField>

                <FormField label="Pagado por" htmlFor="modalPaidBy">
                    <select
                        id="modalPaidBy"
                        className="input"
                        value={paidByMemberId}
                        onChange={(e) => setPaidByMemberId(e.target.value)}
                        disabled={isSubmitting || isLoadingMembers || !hasMembers}
                    >
                        {isLoadingMembers ? (
                            <option>Cargando miembros…</option>
                        ) : !hasMembers ? (
                            <option disabled>No hay miembros elegibles</option>
                        ) : (
                            eligibleMembers.map((m) => (
                                <option key={m.id} value={m.id}>
                                    {m.name}{m.id === defaultPaidByMemberId ? ' (tú)' : ''}
                                </option>
                            ))
                        )}
                    </select>
                    {isCurrentMemberSelected && (
                        <p className="formHint">Por defecto eres tú (miembro vinculado a tu sesión).</p>
                    )}
                    {isCurrentUserOwner && hasMembers && (
                        <p className="formHint">Como owner, puedes registrar gastos para otros miembros.</p>
                    )}
                    {!isLoadingMembers && !hasMembers && (
                        <p className="formHint formHintError">Primero vincula tu cuenta a un miembro para registrar gastos.</p>
                    )}
                </FormField>

                <FormField label="Categoría" htmlFor="modalCategory">
                    <select
                        id="modalCategory"
                        className="input"
                        value={category}
                        onChange={(e) => setCategory(e.target.value)}
                        disabled={isSubmitting || loadingCategories}
                    >
                        {loadingCategories ? (
                            <option>Cargando…</option>
                        ) : (
                            categories.map((c) => (
                                <option key={c.id || c.slug} value={c.id || c.slug}>
                                    {getCategoryIcon(c.slug || c.id)} {c.name}
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
                        <span className="sharedToggleText">Gasto compartido</span>
                        <Tooltip text="Los gastos compartidos se dividen entre miembros según modo del hogar (igual o proporcional)." />
                    </label>
                </div>

                {/* ── Advanced options toggle ── */}
                <button
                    type="button"
                    className="btn btnGhost btnSm advancedToggle"
                    onClick={() => setShowAdvanced((v) => !v)}
                >
                    {showAdvanced ? '▲ Ocultar opciones' : '▼ More options'}
                </button>

                {showAdvanced && (
                    <div className="formStack advancedSection">
                        <FormField label="Método de pago" htmlFor="modalPaymentMethod">
                            <select
                                id="modalPaymentMethod"
                                className="input"
                                value={paymentMethod}
                                onChange={(e) => setPaymentMethod(e.target.value)}
                                disabled={isSubmitting}
                            >
                                <option value="card">💳 Tarjeta</option>
                                <option value="cash">💵 Efectivo</option>
                                <option value="transfer">🏦 Transferencia</option>
                                <option value="voucher">🎟️ Vale</option>
                            </select>
                            {isVoucher && (
                                <p className="formHint">
                                    Los gastos con vale sí se consideran en la conciliación.
                                </p>
                            )}
                        </FormField>

                        {isCardPayment && (
                            <FormField label="Tarjeta" htmlFor="modalCardName">
                                <select
                                    id="modalCardName"
                                    className="input"
                                    value={hasRegisteredCards ? cardId : cardName}
                                    onChange={(e) => {
                                        const selectedValue = e.target.value
                                        if (hasRegisteredCards) {
                                            setCardId(selectedValue)
                                            const selectedCard = cards.find((item) => item.id === selectedValue)
                                            setCardName(selectedCard?.card_name || '')
                                            return
                                        }
                                        setCardName(selectedValue)
                                    }}
                                    disabled={isSubmitting || loadingCards}
                                >
                                    <option value="" disabled>Selecciona tarjeta...</option>
                                    {hasRegisteredCards ? (
                                        cards.map((item) => (
                                            <option key={item.id} value={item.id}>
                                                {item.bank_name} - {item.card_name}
                                            </option>
                                        ))
                                    ) : (
                                        MEXICAN_BANKS.map((bank) => (
                                            <option key={bank.value} value={bank.value}>{bank.label}</option>
                                        ))
                                    )}
                                </select>
                                {hasRegisteredCards && (
                                    <p className="formHint">Usando tarjetas registradas en este hogar.</p>
                                )}
                            </FormField>
                        )}

                        <FormField
                            label={(
                                <>
                                    Expense type
                                    <Tooltip text="Variable es único del periodo. MSI genera cuotas para meses futuros." position="right" />
                                </>
                            )}
                            htmlFor="modalExpenseType"
                        >
                            <select
                                id="modalExpenseType"
                                className="input"
                                value={expenseType}
                                onChange={(e) => setExpenseType(e.target.value)}
                                disabled={isSubmitting}
                            >
                                <option value="variable">📝 Variable</option>
                                <option value="msi">🔒 MSI (meses sin intereses)</option>
                            </select>
                            <p className="formHint">{EXPENSE_TYPE_HINTS[expenseType]}</p>
                            <p className="formHint formHintWarning">Gestiona gastos fijos desde configuración.</p>
                            <button
                                type="button"
                                className="btn btnGhost btnSm"
                                onClick={() => {
                                    onClose()
                                    navigate('/onboarding/fixed-expenses')
                                }}
                            >
                                Ir a gastos fijos
                            </button>
                        </FormField>

                        {isMSI && (
                            <FormField label="Total de cuotas" htmlFor="modalTotalInstallments">
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
                        disabled={!isValid || isSubmitting || isLoadingMembers || !hasMembers || isMutationLocked}
                    >
                        {isSubmitting
                            ? <><span className="spinIcon" aria-hidden>⟳</span> Guardando…</>
                            : 'Agregar gasto'}
                    </button>
                    <button type="button" className="btn btnGhost btnFull" onClick={onClose} disabled={isSubmitting}>
                        Cancelar
                    </button>
                </div>
            </form>
        </Modal>
    )
}
