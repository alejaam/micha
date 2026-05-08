import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { createCard, listCards } from '../api'
import { MEXICAN_BANKS } from '../constants/mexicanBanks'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { useFormField } from '../hooks/useFormField'
import { useSlideDirection } from '../hooks/useSlideDirection'
import {
    AuthCard,
    AuthHeader,
    AuthFormField,
    AuthInput,
    AuthButton,
    AuthBanner,
    AnimatedStep,
} from '../ui/auth'

const ONBOARDING_STEP_PATHS = [
    '/onboarding/household',
    '/onboarding/cards',
    '/onboarding/fixed-expenses',
]

function preferredCardStorageKey(householdId) {
    return `micha_preferred_card_${householdId}`
}

export function OnboardingCardsPage() {
    const { pathname } = useLocation()
    const direction = useSlideDirection(ONBOARDING_STEP_PATHS)
    const { handleProtectedError } = useAuth()
    const { householdId } = useAppShell()
    const navigate = useNavigate()

    const [cards, setCards] = useState([])
    const [selectedCardId, setSelectedCardId] = useState('')
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState('')
    const [message, setMessage] = useState('')
    const [showForm, setShowForm] = useState(true)

    const banco = useFormField(MEXICAN_BANKS[0].value, (v) =>
        v.trim() ? null : 'El banco es obligatorio.',
    )
    const nombreTarjeta = useFormField('', (v) =>
        v.trim() ? null : 'El nombre es obligatorio.',
    )
    const diaCorte = useFormField('15', (v) => {
        if (!v.trim()) return null
        const n = Number(v)
        if (!Number.isInteger(n) || n < 1 || n > 31)
            return 'Ingresa un día entre 1 y 31.'
        return null
    })

    const hasCards = cards.length > 0

    const loadCards = useCallback(async () => {
        if (!householdId) return

        setLoading(true)
        try {
            const data = await listCards({ householdId })
            const items = Array.isArray(data) ? data : []
            setCards(items)

            const preferredCardId =
                localStorage.getItem(
                    preferredCardStorageKey(householdId),
                ) ?? ''
            if (
                preferredCardId &&
                items.some((item) => item.id === preferredCardId)
            ) {
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
        localStorage.setItem(
            preferredCardStorageKey(householdId),
            selectedCardId,
        )
    }, [householdId, selectedCardId])

    const canCreate = useMemo(() => {
        const day = Number(diaCorte.value)
        return (
            banco.value.trim() !== '' &&
            nombreTarjeta.value.trim() !== '' &&
            Number.isInteger(day) &&
            day >= 1 &&
            day <= 31
        )
    }, [banco.value, nombreTarjeta.value, diaCorte.value])

    async function handleCreateCard(e) {
        e.preventDefault()
        if (!householdId || !canCreate) return

        setSaving(true)
        setError('')
        setMessage('')
        try {
            await createCard({
                householdId,
                bankName: banco.value.trim(),
                cardName: nombreTarjeta.value.trim(),
                cutoffDay: Number(diaCorte.value),
            })
            banco.setValue(MEXICAN_BANKS[0].value)
            nombreTarjeta.setValue('')
            diaCorte.setValue('15')
            setMessage('Tarjeta agregada correctamente.')
            setShowForm(false)
            await loadCards()
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setSaving(false)
        }
    }

    function handleContinue() {
        navigate('/onboarding/fixed-expenses', { replace: true })
    }

    if (!householdId) {
        return (
            <AnimatedStep pathname={pathname} direction={direction}>
                <AuthCard>
                    <AuthBanner type="error">
                        No hay un hogar seleccionado. Creá tu hogar primero.
                    </AuthBanner>
                    <AuthButton
                        fullWidth
                        onClick={() =>
                            navigate('/onboarding/household', { replace: true })
                        }
                    >
                        Ir a crear hogar
                    </AuthButton>
                </AuthCard>
            </AnimatedStep>
        )
    }

    return (
        <AnimatedStep pathname={pathname} direction={direction}>
            <AuthCard>
                <AuthHeader
                    eyebrow="Paso 2 de 2"
                    title="Agregar tus tarjetas"
                    subtitle="Creá al menos una tarjeta para usarla al registrar tus primeros gastos."
                />

                {error ? <AuthBanner type="error">{error}</AuthBanner> : null}
                {message && !showForm ? (
                    <AuthBanner type="success">{message}</AuthBanner>
                ) : null}

                {!showForm && hasCards && (
                    <label className="pd-toggleRow" htmlFor="addAnotherCard">
                        <input
                            id="addAnotherCard"
                            type="checkbox"
                            className="pd-toggleCheckbox"
                            checked={showForm}
                            onChange={(e) => {
                                setShowForm(e.target.checked)
                                if (e.target.checked) setMessage('')
                            }}
                        />
                        <span className="pd-toggleLabel">
                            Agregar otra tarjeta
                        </span>
                    </label>
                )}

                {showForm && (
                    <form
                        className="pd-field"
                        onSubmit={handleCreateCard}
                        noValidate
                        aria-label="Agregar tarjeta"
                    >
                        <AuthFormField label="Banco" htmlFor="cardBanco">
                            <select
                                id="cardBanco"
                                className="pd-input"
                                value={banco.value}
                                onChange={(e) =>
                                    banco.setValue(e.target.value)
                                }
                                disabled={saving}
                            >
                                {MEXICAN_BANKS.map((bank) => (
                                    <option key={bank.value} value={bank.value}>
                                        {bank.label}
                                    </option>
                                ))}
                            </select>
                        </AuthFormField>

                        <AuthFormField
                            label="Nombre de la tarjeta"
                            htmlFor="cardNombre"
                            error={nombreTarjeta.error}
                        >
                            <AuthInput
                                id="cardNombre"
                                placeholder="Ej. Platinum"
                                value={nombreTarjeta.value}
                                onChange={(e) =>
                                    nombreTarjeta.setValue(e.target.value)
                                }
                                onBlur={nombreTarjeta.onBlur}
                                disabled={saving}
                                hasError={!!nombreTarjeta.error}
                            />
                        </AuthFormField>

                        <AuthFormField
                            label="Día de corte"
                            htmlFor="cardCorte"
                            error={diaCorte.error}
                        >
                            <AuthInput
                                id="cardCorte"
                                type="number"
                                min="1"
                                max="31"
                                placeholder="15"
                                value={diaCorte.value}
                                onChange={(e) =>
                                    diaCorte.setValue(e.target.value)
                                }
                                onBlur={diaCorte.onBlur}
                                disabled={saving}
                                hasError={!!diaCorte.error}
                            />
                        </AuthFormField>

                        <AuthButton
                            type="submit"
                            fullWidth
                            disabled={!canCreate || saving}
                            busy={saving}
                        >
                            {saving ? 'Guardando…' : 'Guardar tarjeta'}
                        </AuthButton>
                    </form>
                )}

                <div style={{ marginTop: 24 }}>
                    <p
                        style={{
                            fontSize: '0.9375rem',
                            fontWeight: 700,
                            lineHeight: '1.15',
                            color: 'var(--pd-text-primary)',
                            margin: '0 0 12px',
                        }}
                    >
                        Tus tarjetas
                    </p>
                    {loading ? (
                        <p className="pd-hint">Cargando tarjetas…</p>
                    ) : !hasCards ? (
                        <p className="pd-hint">
                            Todavía no tenés tarjetas. Podés agregar una ahora o
                            hacerlo más tarde.
                        </p>
                    ) : (
                        <>
                            <div className="pd-cardList">
                                {cards.map((item) => {
                                    const isSelected =
                                        selectedCardId === item.id
                                    return (
                                        <label
                                            key={item.id}
                                            className={`pd-cardItem${isSelected ? ' pd-cardItemSelected' : ''}`}
                                            htmlFor={`card-${item.id}`}
                                        >
                                            <input
                                                id={`card-${item.id}`}
                                                type="radio"
                                                name="preferred-card"
                                                className="pd-cardRadio"
                                                value={item.id}
                                                checked={isSelected}
                                                onChange={() =>
                                                    setSelectedCardId(item.id)
                                                }
                                            />
                                            <span className="pd-cardLabel">
                                                {item.bank_name} —{' '}
                                                {item.card_name} (corte{' '}
                                                {item.cutoff_day})
                                            </span>
                                        </label>
                                    )
                                })}
                            </div>
                            <p className="pd-hint">
                                La tarjeta seleccionada será la predeterminada
                                al crear gastos.
                            </p>
                        </>
                    )}
                </div>

                <div
                    style={{
                        display: 'flex',
                        gap: 8,
                        marginTop: 24,
                    }}
                >
                    <AuthButton
                        type="button"
                        fullWidth
                        onClick={handleContinue}
                        disabled={!hasCards}
                    >
                        Continuar a gastos fijos
                    </AuthButton>
                </div>
            </AuthCard>
        </AnimatedStep>
    )
}
