import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createCard, deleteCard, listCards } from '../api'
import { MEXICAN_BANKS } from '../constants/mexicanBanks'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

export function OnboardingCardsPage() {
    const { handleProtectedError } = useAuth()
    const { householdId } = useAppShell()
    const navigate = useNavigate()

    const [bankName, setBankName] = useState(MEXICAN_BANKS[0].value)
    const [cardName, setCardName] = useState('')
    const [cutoffDay, setCutoffDay] = useState('15')
    const [cards, setCards] = useState([])
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState('')
    const [message, setMessage] = useState('')
    const [showForm, setShowForm] = useState(false)

    const hasCards = cards.length > 0

    const loadCards = useCallback(async () => {
        if (!householdId) return

        setLoading(true)
        try {
            const data = await listCards({ householdId })
            setCards(Array.isArray(data) ? data : [])
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setLoading(false)
        }
    }, [handleProtectedError, householdId])

    useEffect(() => {
        loadCards()
    }, [loadCards])

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

    async function handleDelete(cardId) {
        if (!confirm('¿Eliminar esta tarjeta? Los gastos registrados con ella no se verán afectados.')) {
            return
        }

        setError('')
        setMessage('')
        try {
            await deleteCard({ cardId, householdId })
            setMessage('Card deleted successfully.')
            await loadCards()
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        }
    }

    function handleContinue() {
        navigate('/onboarding/fixed-expenses', { replace: true })
    }

    function toggleForm() {
        setShowForm((prev) => !prev)
        if (!showForm) setMessage('')
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
                <p className="authEyebrow">Setup step</p>
                <h2 className="authTitle">Add your cards</h2>
                <p className="authMeta">Create at least one card so it is ready when you register your first expense.</p>
            </div>

            {error ? <Banner type="error" floating onDismiss={() => setError('')}>{error}</Banner> : null}
            {message ? <Banner type="ok" floating onDismiss={() => setMessage('')}>{message}</Banner> : null}

            <button type="button" className="btn btnGhost btnSm mt-2" onClick={toggleForm}>
                {showForm ? '− Cancelar' : '+ Agregar tarjeta'}
            </button>

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

            <div className="formSection mt-6">
                <h3 className="sectionTitle">Your cards</h3>
                {loading ? (
                    <p className="text-sm text-dim">Loading cards...</p>
                ) : !hasCards ? (
                    <p className="text-sm text-dim">No cards yet. Add one using the button above.</p>
                ) : (
                    <div className="fixedExpensesTable cardAdminTable">
                        <div className="fixedTableHeader">
                            <span className="cardColBank">Banco</span>
                            <span className="cardColName">Tarjeta</span>
                            <span className="cardColCutoff">Corte</span>
                            <span className="cardColActions">Acción</span>
                        </div>
                        {cards.map((item) => (
                            <div key={item.id} className="fixedTableRow">
                                <span className="cardColBank">{item.bank_name}</span>
                                <span className="cardColName">{item.card_name}</span>
                                <span className="cardColCutoff">{item.cutoff_day}</span>
                                <span className="cardColActions">
                                    <button
                                        type="button"
                                        className="btn btnSm btnGhostDanger"
                                        onClick={() => handleDelete(item.id)}
                                    >
                                        Eliminar
                                    </button>
                                </span>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            <div className="flex gap-4 mt-4">
                <button type="button" className="btn btnPrimary flex-1" onClick={handleContinue} disabled={!hasCards}>Continue to fixed expenses</button>
            </div>
        </section>
    )
}
