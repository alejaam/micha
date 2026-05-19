import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createCard, deleteCard, listCards } from '../api'
import { MEXICAN_BANKS } from '../constants/mexicanBanks'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

const SETTLEMENT_HINTS = {
    equal: 'Cada miembro paga la misma parte, sin importar ingresos.',
    proportional: 'Los miembros que ganan más contribuyen con una mayor parte de los gastos.',
}

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
    const [showForm, setShowForm] = useState(false)

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
            setMessage('Tarjeta añadida correctamente.')
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
                <Banner type="error">No hay hogar seleccionado. Crea tu hogar primero.</Banner>
                <button className="btn u-mt-4" onClick={() => navigate('/onboarding/household', { replace: true })}>Ir a configuración del hogar</button>
            </div>
        )
    }

    return (
        <section className="card onboardingCard" aria-label="Configura tus tarjetas">
            <div className="onboardingHeader">
                <p className="authEyebrow">Paso de configuración</p>
                <h2 className="authTitle">Añade tus tarjetas</h2>
                <p className="authMeta">Crea al menos una tarjeta para tenerla lista cuando registres tu primer gasto.</p>
            </div>

            {error ? <Banner type="error" floating onDismiss={() => setError('')}>{error}</Banner> : null}
            {message ? <Banner type="ok" floating onDismiss={() => setMessage('')}>{message}</Banner> : null}

            <button type="button" className="btn btnGhost btnSm u-mt-2" onClick={toggleForm}>
                {showForm ? '− Cancelar' : '+ Agregar tarjeta'}
            </button>

            {showForm && (
                <form className="formStack u-mt-4" onSubmit={handleCreateCard}>
                    <FormField label="Banco" htmlFor="onboardingBankName">
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

                <FormField label="Nombre de la tarjeta" htmlFor="onboardingCardName">
                    <input
                        id="onboardingCardName"
                        className="input"
                        value={cardName}
                        onChange={(e) => setCardName(e.target.value)}
                        placeholder="Ej. Platino"
                        disabled={saving}
                    />
                </FormField>

                <FormField label="Día de corte" htmlFor="onboardingCutoffDay">
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

                <button type="submit" className="btn btnPrimary u-w-full" disabled={!canCreate || saving}>
                    {saving ? 'Guardando...' : 'Guardar tarjeta'}
                </button>
            </form>
            )}

            <div className="formSection u-mt-8">
                <h3 className="sectionTitle">Tus tarjetas</h3>
                {loading ? (
                    <p className="u-text-sm u-text-dim">Cargando tarjetas...</p>
                ) : !hasCards ? (
                    <p className="u-text-sm u-text-dim">Aún no hay tarjetas. Puedes añadir una ahora o saltar este paso.</p>
                ) : (
                    <div className="formStack">
                        {cards.filter(Boolean).map((item) => (
                            <div key={item?.id} className="u-flex u-items-center u-gap-2">
                                <label className="sharedToggleLabel u-flex-1" htmlFor={`preferred-card-${item?.id}`}>
                                    <input
                                        id={`preferred-card-${item?.id}`}
                                        type="radio"
                                        name="preferred-card"
                                        value={item?.id}
                                        checked={selectedCardId === item?.id}
                                        onChange={() => setSelectedCardId(item?.id)}
                                    />
                                    <span className="sharedToggleText">{item?.bank_name} - {item?.card_name} (corte {item?.cutoff_day})</span>
                                </label>
                                <button
                                    type="button"
                                    className="btn btnSm btnGhostDanger"
                                    onClick={() => handleDelete(item?.id)}
                                    title="Eliminar tarjeta"
                                >
                                    ✕
                                </button>
                            </div>
                        ))}
                        <p className="formHint">La tarjeta seleccionada será la predeterminada al crear gastos.</p>
                    </div>
                )}
            </div>

            <div className="u-flex u-gap-4 u-mt-4">
                <button type="button" className="btn btnPrimary u-flex-1" onClick={handleContinue} disabled={!hasCards}>Continuar a gastos fijos</button>
            </div>
        </section>
    )
}
