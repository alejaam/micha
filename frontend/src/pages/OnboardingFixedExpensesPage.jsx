import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createExpense } from '../api'
import { FIXED_EXPENSES_COMMON } from '../constants/fixedExpenses'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { useCurrentMember } from '../hooks/useCurrentMember'
import { useMembers } from '../hooks/useMembers'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'
import { dollarsToCents, sanitizeAmountInput } from '../utils'

export function OnboardingFixedExpensesPage() {
    const { handleProtectedError, user, isAuthenticated } = useAuth()
    const { householdId, selectedHousehold } = useAppShell()
    const navigate = useNavigate()

    const { members, loadingMembers } = useMembers({
        isAuthenticated,
        householdId,
        handleProtectedError,
    })

    const currentMember = useCurrentMember(members)

    const [selected, setSelected] = useState({})
    const [amounts, setAmounts] = useState({})
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState('')
    const [message, setMessage] = useState('')

    const linkedMembers = useMemo(
        () => members.filter((m) => m.user_id && String(m.user_id).trim() !== ''),
        [members],
    )

    const adminMember = linkedMembers[0] ?? null
    const currency = selectedHousehold?.currency || 'MXN'

    const canSubmit = useMemo(() => {
        const ids = Object.keys(selected).filter((id) => selected[id])
        if (ids.length === 0) return true

        return ids.every((id) => {
            const cents = dollarsToCents(amounts[id] || '')
            return cents !== null && cents > 0
        })
    }, [amounts, selected])

    function toggleItem(itemId) {
        setSelected((prev) => {
            const next = { ...prev, [itemId]: !prev[itemId] }
            if (!next[itemId]) {
                setAmounts((current) => {
                    const copy = { ...current }
                    delete copy[itemId]
                    return copy
                })
            }
            return next
        })
    }

    function handleSkip() {
        navigate('/', { replace: true })
    }

    async function handleSubmit(e) {
        e.preventDefault()
        if (!householdId) return

        setSaving(true)
        setError('')
        setMessage('')

        try {
            const selectedItems = FIXED_EXPENSES_COMMON.filter((item) => selected[item.id])
            if (selectedItems.length === 0) {
                navigate('/', { replace: true })
                return
            }

            const paidByMemberId = currentMember?.id || adminMember?.id || ''
            if (!paidByMemberId) {
                throw new Error('no linked member available to register fixed expenses')
            }

            for (const item of selectedItems) {
                const amountCents = dollarsToCents(amounts[item.id] || '')
                if (amountCents === null || amountCents <= 0) {
                    throw new Error(`invalid amount for ${item.label}`)
                }

                await createExpense({
                    householdId,
                    paidByMemberId,
                    amountCents,
                    description: item.label,
                    isShared: true,
                    currency,
                    paymentMethod: 'transfer',
                    expenseType: 'fixed',
                    category: item.id,
                    totalInstallments: 0,
                })
            }

            setMessage('Fixed expenses saved. Redirecting...')
            navigate('/', { replace: true })
        } catch (err) {
            if (!handleProtectedError(err)) {
                setError(err.message || 'failed to register fixed expenses')
            }
        } finally {
            setSaving(false)
        }
    }

    if (!householdId) {
        return (
            <section className="card onboardingCard" aria-label="Set up fixed expenses">
                <Banner type="error">No household selected. Create your household first.</Banner>
                <button type="button" className="btn mt-4" onClick={() => navigate('/onboarding/household', { replace: true })}>
                    Go to household setup
                </button>
            </section>
        )
    }

    return (
        <section className="card onboardingCard" aria-label="Set up fixed expenses">
            <div className="onboardingHeader">
                <p className="authEyebrow">Optional step</p>
                <h2 className="authTitle">Add common fixed expenses</h2>
                <p className="authMeta">Select what applies now and set an estimated monthly amount. You can skip and add them later.</p>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}
            {message ? <Banner type="ok">{message}</Banner> : null}

            {loadingMembers ? (
                <p className="text-sm text-dim">Loading members...</p>
            ) : (
                <p className="formHint">
                    Expenses will be registered under <strong>{currentMember?.name || adminMember?.name || user?.email || 'the admin member'}</strong>.
                </p>
            )}

            <form className="formStack mt-4" onSubmit={handleSubmit}>
                {FIXED_EXPENSES_COMMON.map((item) => {
                    const checked = !!selected[item.id]
                    return (
                        <div key={item.id} className="sharedToggle">
                            <label className="sharedToggleLabel" htmlFor={`fixed-${item.id}`}>
                                <input
                                    id={`fixed-${item.id}`}
                                    type="checkbox"
                                    checked={checked}
                                    onChange={() => toggleItem(item.id)}
                                    disabled={saving}
                                />
                                <span className="sharedToggleText">{item.label}</span>
                            </label>
                            {checked && (
                                <FormField label={`Monthly amount (${currency})`} htmlFor={`fixed-amount-${item.id}`}>
                                    <div className="inputWrap">
                                        <span className="inputPrefix" aria-hidden>$</span>
                                        <input
                                            id={`fixed-amount-${item.id}`}
                                            className="input inputWithPrefix"
                                            inputMode="decimal"
                                            placeholder={item.placeholder}
                                            value={amounts[item.id] || ''}
                                            onChange={(e) => setAmounts((prev) => ({ ...prev, [item.id]: sanitizeAmountInput(e.target.value) }))}
                                            disabled={saving}
                                        />
                                    </div>
                                </FormField>
                            )}
                        </div>
                    )
                })}

                <p className="formHint formHintWarning">This step is optional. If skipped, you can still add fixed expenses from the expense modal later.</p>

                <div className="flex gap-4 mt-4">
                    <button type="button" className="btn flex-1" onClick={handleSkip} disabled={saving}>
                        Skip for now
                    </button>
                    <button type="submit" className="btn btnPrimary flex-1" disabled={saving || !canSubmit || loadingMembers}>
                        {saving ? 'Saving...' : 'Save and continue'}
                    </button>
                </div>
            </form>
        </section>
    )
}
