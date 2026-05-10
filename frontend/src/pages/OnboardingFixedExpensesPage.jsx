import { useCallback, useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { createRecurringExpense, deleteRecurringExpense, initializePeriod, listRecurringExpenses, updateRecurringExpense } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { dollarsToCents, formatCurrency, sanitizeAmountInput } from '../utils'

const CATEGORY_OPTIONS = [
    { value: 'rent', label: 'Renta' },
    { value: 'auto', label: 'Auto' },
    { value: 'streaming', label: 'Streaming / Servicios' },
    { value: 'food', label: 'Comida' },
    { value: 'personal', label: 'Personal' },
    { value: 'savings', label: 'Ahorros' },
    { value: 'other', label: 'Otro' },
]

const CATEGORY_LABEL_MAP = Object.fromEntries(CATEGORY_OPTIONS.map((c) => [c.value, c.label]))

const FIXED_EXPENSE_OPTIONS = [
    { key: 'rent', label: 'Rent', category: 'rent' },
    { key: 'internet', label: 'Internet', category: 'other' },
    { key: 'subscriptions', label: 'Subscriptions', category: 'streaming' },
    { key: 'auto', label: 'Auto', category: 'auto' },
    { key: 'mortgage', label: 'Mortgage', category: 'rent' },
    { key: 'other', label: 'Other', category: 'other' },
]

function todayDateOnly() {
    return new Date().toISOString().slice(0, 10)
}

export function OnboardingFixedExpensesPage() {
    const { householdId } = useAppShell()
    const { handleProtectedError } = useAuth()
    const navigate = useNavigate()
    const location = useLocation()

    // Creation form state
    const [selected, setSelected] = useState({})
    const [amountByKey, setAmountByKey] = useState({})
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState('')
    const [message, setMessage] = useState('')

    // Admin panel state
    const [recurringItems, setRecurringItems] = useState([])
    const [loading, setLoading] = useState(false)
    const [editingId, setEditingId] = useState(null)
    const [editForm, setEditForm] = useState({ description: '', amount: '', category: '' })

    const isOnboarding = location.pathname.startsWith('/onboarding/')

    // ── Load existing recurring expenses ──────────────────────────────────

    const loadRecurringExpenses = useCallback(async () => {
        if (!householdId) return
        setLoading(true)
        try {
            const data = await listRecurringExpenses({ householdId })
            setRecurringItems(Array.isArray(data) ? data : [])
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setLoading(false)
        }
    }, [handleProtectedError, householdId])

    useEffect(() => {
        loadRecurringExpenses()
    }, [loadRecurringExpenses])

    // Total monthly impact
    const totalMonthlyCents = useMemo(
        () => recurringItems.reduce((sum, item) => sum + (item.amount_cents || 0), 0),
        [recurringItems],
    )

    // ── Creation form logic ────────────────────────────────────────────────

    const selectedKeys = useMemo(
        () => FIXED_EXPENSE_OPTIONS.filter((item) => selected[item.key]).map((item) => item.key),
        [selected],
    )

    function toggleOption(key) {
        setSelected((prev) => ({ ...prev, [key]: !prev[key] }))
    }

    function handleAmountChange(key, value) {
        setAmountByKey((prev) => ({ ...prev, [key]: sanitizeAmountInput(value) }))
    }

    const hasValidSelection = useMemo(() => {
        if (selectedKeys.length === 0) return false
        return selectedKeys.every((key) => dollarsToCents(amountByKey[key] ?? '') !== null)
    }, [selectedKeys, amountByKey])

    // ── Edit handlers ───────────────────────────────────────────────────────

    function startEdit(item) {
        setEditingId(item.id)
        setEditForm({
            description: item.description || '',
            amount: (item.amount_cents / 100).toFixed(2),
            category: item.category_id || item.category || 'other',
        })
    }

    function cancelEdit() {
        setEditingId(null)
        setEditForm({ description: '', amount: '', category: '' })
    }

    function handleEditFormChange(field, value) {
        setEditForm((prev) => ({ ...prev, [field]: value }))
    }

    async function handleSaveEdit(itemId) {
        setError('')
        setMessage('')
        try {
            const amountCents = dollarsToCents(editForm.amount)
            if (amountCents === null) {
                setError('Monto inválido')
                return
            }
            await updateRecurringExpense({
                recurringExpenseId: itemId,
                description: editForm.description,
                amountCents,
                category: editForm.category,
            })
            setMessage('Gasto fijo actualizado.')
            cancelEdit()
            await loadRecurringExpenses()
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        }
    }

    async function handleDelete(itemId) {
        if (!confirm('¿Eliminar este gasto fijo?')) return

        setError('')
        setMessage('')
        try {
            await deleteRecurringExpense({ recurringExpenseId: itemId })
            setMessage('Gasto fijo eliminado.')
            await loadRecurringExpenses()
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        }
    }

    // ── Save new fixed expenses (creation form) ────────────────────────────

    async function handleSave() {
        if (!householdId || !hasValidSelection) return

        setSaving(true)
        setError('')
        setMessage('')
        try {
            const startDate = todayDateOnly()
            for (const key of selectedKeys) {
                const config = FIXED_EXPENSE_OPTIONS.find((item) => item.key === key)
                if (!config) continue
                const amountCents = dollarsToCents(amountByKey[key] ?? '')
                if (amountCents === null) continue

                await createRecurringExpense({
                    householdId,
                    paidByMemberId: '',
                    isAgnostic: true,
                    amountCents,
                    description: config.label,
                    category: config.category,
                    expenseType: 'fixed',
                    recurrencePattern: 'monthly',
                    startDate,
                })
            }

            // Refresh the list
            await loadRecurringExpenses()

            if (isOnboarding) {
                try {
                    await initializePeriod({ householdId })
                    const monthNames = ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre']
                    const currentMonth = monthNames[new Date().getMonth()]
                    setMessage(`Tu período de ${currentMonth} ha comenzado automáticamente. ¡Bienvenido a micha!`)
                } catch (initErr) {
                    if (initErr.message?.includes('already has periods')) {
                        // Period already exists — not an error, just proceed
                    } else {
                        setError(initErr.message)
                    }
                }

                navigate('/', { replace: true })
            } else {
                // Clear form and show success message
                setSelected({})
                setAmountByKey({})
                setMessage('Gastos fijos guardados.')
            }
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setSaving(false)
        }
    }

    // ── Render ─────────────────────────────────────────────────────────────

    if (!householdId) {
        return (
            <section className="card onboardingCard" aria-label="Fixed expense setup">
                <Banner type="error">No household selected. Complete household setup first.</Banner>
                <button
                    type="button"
                    className="btn mt-4"
                    onClick={() => navigate('/onboarding/household', { replace: true })}
                >
                    Go to household setup
                </button>
            </section>
        )
    }

    return (
        <section className="card onboardingCard" aria-label="Fixed expenses setup">
            <div className="onboardingHeader">
                <p className="authEyebrow">{isOnboarding ? 'Optional setup' : 'Administración'}</p>
                <h2 className="authTitle">{isOnboarding ? 'Add household fixed expenses' : 'Gastos fijos'}</h2>
                <p className="authMeta">
                    {isOnboarding
                        ? 'Choose the fixed expenses you want to track monthly. These are shared, household-level templates.'
                        : 'Administra tus gastos fijos recurrentes: edita montos, categorías o elimina gastos existentes.'}
                </p>
            </div>

            {error ? <Banner type="error" floating onDismiss={() => setError('')}>{error}</Banner> : null}
            {message ? <Banner type="ok" floating onDismiss={() => setMessage('')}>{message}</Banner> : null}

            {/* ── Summary banner ──────────────────────────────────────────── */}
            {recurringItems.length > 0 && (
                <div className="summaryBanner">
                    <span className="summaryBannerLabel">Tus gastos fijos suman</span>
                    <span className="summaryBannerValue">{formatCurrency(totalMonthlyCents, 'MXN')}</span>
                    <span className="summaryBannerLabel">este mes</span>
                </div>
            )}

            {/* ── Loading state ───────────────────────────────────────────── */}
            {loading && (
                <p className="text-sm text-dim mt-4">Cargando gastos fijos...</p>
            )}

            {/* ── Table of existing items ─────────────────────────────────── */}
            {!loading && recurringItems.length > 0 && (
                <div className="fixedExpensesTable fixedAdminTable mt-4">
                    <div className="fixedTableHeader">
                        <span className="fixedColDesc">Descripción</span>
                        <span className="fixedColCategory">Categoría</span>
                        <span className="fixedColAmount">Monto</span>
                        <span className="fixedColActions">Acciones</span>
                    </div>
                    {recurringItems.map((item) => {
                        const isEditing = editingId === item.id
                        return (
                            <div key={item.id} className="fixedTableRow">
                                {isEditing ? (
                                    <>
                                        <span className="fixedColDesc">
                                            <input
                                                className="input inputSm"
                                                value={editForm.description}
                                                onChange={(e) => handleEditFormChange('description', e.target.value)}
                                            />
                                        </span>
                                        <span className="fixedColCategory">
                                            <select
                                                className="input inputSm"
                                                value={editForm.category}
                                                onChange={(e) => handleEditFormChange('category', e.target.value)}
                                            >
                                                {CATEGORY_OPTIONS.map((opt) => (
                                                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                                                ))}
                                            </select>
                                        </span>
                                        <span className="fixedColAmount">
                                            <div className="inputWrap">
                                                <span className="inputPrefix" aria-hidden>$</span>
                                                <input
                                                    className="input inputWithPrefix inputSm"
                                                    inputMode="decimal"
                                                    value={editForm.amount}
                                                    onChange={(e) => handleEditFormChange('amount', sanitizeAmountInput(e.target.value))}
                                                />
                                            </div>
                                        </span>
                                        <span className="fixedColActions">
                                            <button
                                                type="button"
                                                className="btn btnSm btnPrimary"
                                                onClick={() => handleSaveEdit(item.id)}
                                            >
                                                Guardar
                                            </button>
                                            <button
                                                type="button"
                                                className="btn btnSm btnGhost"
                                                style={{ marginLeft: 4 }}
                                                onClick={cancelEdit}
                                            >
                                                Cancelar
                                            </button>
                                        </span>
                                    </>
                                ) : (
                                    <>
                                        <span className="fixedColDesc">{item.description}</span>
                                        <span className="fixedColCategory">{CATEGORY_LABEL_MAP[item.category_id || item.category] || item.category_id || item.category || '—'}</span>
                                        <span className="fixedColAmount">{formatCurrency(item.amount_cents, 'MXN')}</span>
                                        <span className="fixedColActions">
                                            <button
                                                type="button"
                                                className="btn btnSm btnGhost"
                                                onClick={() => startEdit(item)}
                                            >
                                                Editar
                                            </button>
                                            <button
                                                type="button"
                                                className="btn btnSm btnGhostDanger"
                                                style={{ marginLeft: 4 }}
                                                onClick={() => handleDelete(item.id)}
                                            >
                                                Eliminar
                                            </button>
                                        </span>
                                    </>
                                )}
                            </div>
                        )
                    })}
                </div>
            )}

            {/* ── Empty state ─────────────────────────────────────────────── */}
            {!loading && recurringItems.length === 0 && (
                <div className="emptyState mt-4">
                    <p className="emptyTitle">Sin gastos fijos aún</p>
                    <p className="emptyHint">Agrega gastos fijos usando el formulario de abajo.</p>
                </div>
            )}

            {/* ── Quick-add form ──────────────────────────────────────────── */}
            <div className="mt-6">
                <h3 className="sectionTitle">Agregar nuevo gasto fijo</h3>
                <div className="formStack mt-2">
                    {FIXED_EXPENSE_OPTIONS.map((item) => {
                        const isChecked = !!selected[item.key]
                        return (
                            <div key={item.key} className="formSection">
                                <label className="sharedToggleLabel" htmlFor={`fixed-${item.key}`}>
                                    <input
                                        id={`fixed-${item.key}`}
                                        type="checkbox"
                                        checked={isChecked}
                                        onChange={() => toggleOption(item.key)}
                                        disabled={saving}
                                    />
                                    <span className="sharedToggleText">{item.label}</span>
                                </label>

                                {isChecked && (
                                    <div className="inputWrap mt-2">
                                        <span className="inputPrefix" aria-hidden>$</span>
                                        <input
                                            className="input inputWithPrefix"
                                            inputMode="decimal"
                                            placeholder="0.00"
                                            value={amountByKey[item.key] ?? ''}
                                            onChange={(e) => handleAmountChange(item.key, e.target.value)}
                                            disabled={saving}
                                        />
                                    </div>
                                )}
                            </div>
                        )
                    })}
                </div>
            </div>

            {/* ── Actions ────────────────────────────────────────────────── */}
            <div className="flex gap-4 mt-6">
                {isOnboarding ? (
                    <>
                        <button
                            type="button"
                            className="btn flex-1"
                            onClick={() => navigate('/', { replace: true })}
                            disabled={saving}
                        >
                            Skip for now
                        </button>
                        <button
                            type="button"
                            className="btn btnPrimary flex-1"
                            onClick={handleSave}
                            disabled={saving || !hasValidSelection}
                        >
                            {saving ? 'Saving...' : 'Save and continue'}
                        </button>
                    </>
                ) : (
                    selectedKeys.length > 0 && (
                        <button
                            type="button"
                            className="btn btnPrimary"
                            onClick={handleSave}
                            disabled={saving || !hasValidSelection}
                        >
                            {saving ? 'Guardando...' : 'Guardar gastos fijos'}
                        </button>
                    )
                )}
            </div>
        </section>
    )
}
