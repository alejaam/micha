import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createRecurringExpense } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { dollarsToCents, sanitizeAmountInput } from '../utils'

const FIXED_EXPENSE_OPTIONS = [
    { key: 'rent', label: 'Renta', category: 'rent' },
    { key: 'internet', label: 'Internet', category: 'other' },
    { key: 'subscriptions', label: 'Suscripciones', category: 'streaming' },
    { key: 'auto', label: 'Auto', category: 'auto' },
    { key: 'mortgage', label: 'Hipoteca', category: 'rent' },
    { key: 'other', label: 'Otro', category: 'other' },
]

function todayDateOnly() {
    return new Date().toISOString().slice(0, 10)
}

export function OnboardingFixedExpensesPage() {
    const { householdId } = useAppShell()
    const { handleProtectedError } = useAuth()
    const navigate = useNavigate()

    const [selected, setSelected] = useState({})
    const [amountByKey, setAmountByKey] = useState({})
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState('')
    const [message, setMessage] = useState('')

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

            setMessage('Gastos fijos guardados.')
            navigate('/', { replace: true })
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setSaving(false)
        }
    }

    if (!householdId) {
        return (
            <section className="card onboardingCard" aria-label="Fixed expense setup">
                <Banner type="error">No hay hogar seleccionado. Completa la configuración del hogar primero.</Banner>
                <button
                    type="button"
                    className="btn u-mt-4"
                    onClick={() => navigate('/onboarding/household', { replace: true })}
                >
                    Ir a configuración del hogar
                </button>
            </section>
        )
    }

    return (
        <section className="card onboardingCard" aria-label="Configuración de gastos fijos">
            <div className="onboardingHeader">
                <p className="authEyebrow">Configuración opcional</p>
                <h2 className="authTitle">Añade gastos fijos del hogar</h2>
                <p className="authMeta">
                    Elige los gastos fijos que quieras registrar mensualmente. Son plantillas compartidas a nivel del hogar.
                </p>
            </div>

            {error ? <Banner type="error" floating onDismiss={() => setError('')}>{error}</Banner> : null}
            {message ? <Banner type="ok" floating onDismiss={() => setMessage('')}>{message}</Banner> : null}

            <div className="formStack u-mt-4">
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
                                <div className="inputWrap u-mt-2">
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

            <div className="u-flex u-gap-4 u-mt-6">
                <button
                    type="button"
                    className="btn u-flex-1"
                    onClick={() => navigate('/', { replace: true })}
                    disabled={saving}
                >
                    Saltar por ahora
                </button>
                <button
                    type="button"
                    className="btn btnPrimary u-flex-1"
                    onClick={handleSave}
                    disabled={saving || !hasValidSelection}
                >
                    {saving ? 'Guardando...' : 'Guardar y continuar'}
                </button>
            </div>
        </section>
    )
}
