import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createCard, listCards } from '../api'
import { MEXICAN_BANKS } from '../constants/mexicanBanks'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

function preferredCardStorageKey(householdId) {
    return `micha_preferred_card_${householdId}`
}

export function OnboardingCardsPage() {
    const { handleProtectedError } = useAuth()
    const { householdId } = useAppShell()
    const navigate = useNavigate()

    const [bankName, setBankName] = useState(MEXICAN_BANKS[0].value)
    const [cardName, setCardName] = useState('')
    const [cutoffDay, setCutoffDay] = useState('15')
    const [cards, setCards] = useState([])
    const [selectedCardId, setSelectedCardId] = useState('')
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState('')
    const [message, setMessage] = useState('')
    const [showForm, setShowForm] = useState(true)

    const hasCards = cards.length > 0

    const loadCards = useCallback(async () => {
        if (!householdId) return

        setLoading(true)
        try {
            const data = await listCards({ householdId })
            const items = Array.isArray(data) ? data : []
            setCards(items)

            const preferredCardId = localStorage.getItem(preferredCardStorageKey(householdId)) ?? ''
            if (preferredCardId && items.some((item) => item.id === preferredCardId)) {
                setSelectedCardId(preferredCardId)
            } else if (items.length > 0) {
                setSelectedCardId(items[0].id)
            } else {
                setSelectedCardId('')
            }
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setLoading(false)
        }
    }, [handleProtectedError, householdId])

    useEffect(() => {
        loadCards()
    }, [loadCards])

    useEffect(() => {
        if (!householdId || !selectedCardId) return
        localStorage.setItem(preferredCardStorageKey(householdId), selectedCardId)
    }, [householdId, selectedCardId])

    const canCreate = useMemo(() => {
        const day = Number(cutoffDay)
        return bankName.trim() !== '' && cardName.trim() !== '' && Number.isInteger(day) && day >= 1 && day <= 31
    }, [bankName, cardName, cutoffDay])

    async function handleCreateCard(e) {
        e.preventDefault()
        if (!householdId || !canCreate) return

        setSaving(true)
        setError('')
        setMessage('')
        try {
            await createCard({
                householdId,
                bankName: bankName.trim(),
                cardName: cardName.trim(),
                cutoffDay: Number(cutoffDay),
            })
            setBankName(MEXICAN_BANKS[0].value)
            setCardName('')
            setCutoffDay('15')
            setMessage('Card added successfully.')
            setShowForm(false)
            await loadCards()
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setSaving(false)
        }
    }

    function handleContinue() {
        navigate('/', { replace: true })
    }

    if (!householdId) {
        return (
            <div className="card">
                <Banner type="error">No household selected. Create your household first.</Banner>
                <button className="btn mt-4" onClick={() => navigate('/onboarding/household', { replace: true })}>Go to household setup</button>
            </div>
        )
    }

    return (
        <section className="card onboardingCard" aria-label="Set up your cards">
            <div className="onboardingHeader">
                <p className="authEyebrow">Final step</p>
                <h2 className="authTitle">Add your cards</h2>
                <p className="authMeta">Create at least one card so it is ready when you register your first expense.</p>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}
            {message && !showForm ? <Banner type="ok">{message}</Banner> : null}

            {!showForm && hasCards && (
                <div className="card mt-4 p-4 border border-dim rounded-md bg-secondary">
                    <label className="sharedToggleLabel mb-0 flex items-center gap-2 cursor-pointer" htmlFor="addAnotherCard">
                        <input
                            id="addAnotherCard"
                            type="checkbox"
                            className="w-5 h-5 accent-primary"
                            checked={showForm}
                            onChange={(e) => {
                                setShowForm(e.target.checked)
                                if (e.target.checked) setMessage('')
                            }}
                        />
                        <span className="font-medium text-primary">Add another card</span>
                    </label>
                </div>
            )}

            {showForm && (
                <form className="formStack mt-4" onSubmit={handleCreateCard}>
                    <FormField label="Bank" htmlFor="onboardingBankName">
                    <select
                        id="onboardingBankName"
                        className="input"
                        value={bankName}
                        onChange={(e) => setBankName(e.target.value)}
                        disabled={saving}
                    >
                        {MEXICAN_BANKS.map((bank) => (
                            <option key={bank.value} value={bank.value}>{bank.label}</option>
                        ))}
                    </select>
                </FormField>

                <FormField label="Card name" htmlFor="onboardingCardName">
                    <input
                        id="onboardingCardName"
                        className="input"
                        value={cardName}
                        onChange={(e) => setCardName(e.target.value)}
                        placeholder="e.g. Platinum"
                        disabled={saving}
                    />
                </FormField>

                <FormField label="Cutoff day" htmlFor="onboardingCutoffDay">
                    <input
                        id="onboardingCutoffDay"
                        className="input"
                        type="number"
                        min="1"
                        max="31"
                        value={cutoffDay}
                        onChange={(e) => setCutoffDay(e.target.value)}
                        disabled={saving}
                    />
                </FormField>

                <button type="submit" className="btn btnPrimary w-full" disabled={!canCreate || saving}>
                    {saving ? 'Adding...' : 'Save card'}
                </button>
            </form>
            )}

            <div className="formSection mt-8">
                <h3 className="sectionTitle">Your cards</h3>
                {loading ? (
                    <p className="text-sm text-dim">Loading cards...</p>
                ) : !hasCards ? (
                    <p className="text-sm text-dim">No cards yet. You can add one now or skip and do it later.</p>
                ) : (
                    <div className="formStack">
                        {cards.map((item) => (
                            <label key={item.id} className="sharedToggleLabel" htmlFor={`preferred-card-${item.id}`}>
                                <input
                                    id={`preferred-card-${item.id}`}
                                    type="radio"
                                    name="preferred-card"
                                    value={item.id}
                                    checked={selectedCardId === item.id}
                                    onChange={() => setSelectedCardId(item.id)}
                                />
                                <span className="sharedToggleText">{item.bank_name} - {item.card_name} (cutoff {item.cutoff_day})</span>
                            </label>
                        ))}
                        <p className="formHint">Selected card will be preselected when creating expenses.</p>
                    </div>
                )}
            </div>

            <div className="flex gap-4 mt-4">
                <button type="button" className="btn btnPrimary flex-1" onClick={handleContinue} disabled={!hasCards}>Continue to dashboard</button>
            </div>
        </section>
    )
}
