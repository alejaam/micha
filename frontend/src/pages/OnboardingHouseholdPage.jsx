import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createHousehold } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

const CURRENCIES = [
    { code: 'MXN', label: '🇲🇽 MXN — Peso Mexicano' },
    { code: 'USD', label: '🇺🇸 USD — Dólar Estadounidense' },
    { code: 'EUR', label: '🇪🇺 EUR — Euro' },
    { code: 'COP', label: '🇨🇴 COP — Peso Colombiano' },
    { code: 'ARS', label: '🇦🇷 ARS — Peso Argentino' },
    { code: 'CLP', label: '🇨🇱 CLP — Peso Chileno' },
    { code: 'PEN', label: '🇵🇪 PEN — Sol Peruano' },
    { code: 'BRL', label: '🇧🇷 BRL — Real Brasileño' },
]

const SETTLEMENT_HINTS = {
    equal: 'Cada miembro paga la misma parte, sin importar ingresos.',
    proportional: 'Los miembros que ganan más contribuyen con una mayor parte de los gastos.',
}

export function OnboardingHouseholdPage() {
    const { handleProtectedError } = useAuth()
    const { setHouseholdId, loadHouseholds } = useAppShell()
    const navigate = useNavigate()

    // Household state
    const [hhName, setHhName] = useState('')
    const [settlementMode, setSettlementMode] = useState('equal')
    const [currency, setCurrency] = useState('MXN')
    const [closingDay, setClosingDay] = useState(15)
    const [periodFrequency, setPeriodFrequency] = useState('monthly')

    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    async function handleSubmit(e) {
        e.preventDefault()
        if (!hhName.trim()) return
        
        setBusy(true)
        setError('')
        
        try {
            // 1. Create household
            const hhOut = await createHousehold({
                name: hhName.trim(),
                settlementMode,
                currency,
                closingDay: Number(closingDay),
                periodFrequency,
            })
            
            const createdHouseholdId = hhOut?.household_id ?? hhOut?.id ?? ''
            if (!createdHouseholdId) {
                throw new Error('El hogar fue creado pero no se devolvió el ID')
            }

            // Keep the new ID locally
            setHouseholdId(createdHouseholdId)

            // 2. Refresh households list
            await loadHouseholds()

            // 3. Continue to member onboarding
            navigate('/onboarding/member', { replace: true })
            
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <section className="card onboardingCard" aria-label="Crea tu hogar">
            <div className="onboardingHeader">
                <p className="authEyebrow">Primeros pasos</p>
                <h2 className="authTitle">Configura tu hogar</h2>
                <p className="authMeta">Un hogar agrupa todos los gastos compartidos y los miembros.</p>
            </div>

            <Banner type="info">
                Bienvenido a micha. Para comenzar, necesitas crear tu primer hogar. Este paso es obligatorio.
            </Banner>

            {error ? <Banner type="error" floating onDismiss={() => setError('')}>{error}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit}>
                <div className="formSection">
                    <h3 className="sectionTitle">Detalles del hogar</h3>
                    <FormField label="Nombre del hogar" htmlFor="hhName">
                        <input
                            id="hhName"
                            className="input"
                            placeholder="Ej. Casa Familia"
                            value={hhName}
                            onChange={(e) => setHhName(e.target.value)}
                            disabled={busy}
                            autoFocus
                        />
                    </FormField>
                    <FormField label="Modo de liquidación" htmlFor="hhMode">
                        <select
                            id="hhMode"
                            className="input"
                            value={settlementMode}
                            onChange={(e) => setSettlementMode(e.target.value)}
                            disabled={busy}
                        >
                            <option value="equal">Dividir equitativamente</option>
                            <option value="proportional">Proporcional al salario</option>
                        </select>
                        <p className="formHint">{SETTLEMENT_HINTS[settlementMode]}</p>
                    </FormField>
                    <FormField label="Moneda" htmlFor="hhCurrency">
                        <select
                            id="hhCurrency"
                            className="input"
                            value={currency}
                            onChange={(e) => setCurrency(e.target.value)}
                            disabled={busy}
                        >
                            {CURRENCIES.map((c) => (
                                <option key={c.code} value={c.code}>{c.label}</option>
                            ))}
                        </select>
                    </FormField>

                    <FormField label="Día de cierre" htmlFor="hhClosingDay">
                        <input
                            id="hhClosingDay"
                            className="input"
                            type="number"
                            min="1"
                            max="31"
                            value={closingDay}
                            onChange={(e) => setClosingDay(e.target.value)}
                            disabled={busy}
                        />
                        <p className="formHint">Día del mes en que se cierra el periodo.</p>
                    </FormField>

                    <FormField label="Frecuencia del periodo" htmlFor="hhFrequency">
                        <select
                            id="hhFrequency"
                            className="input"
                            value={periodFrequency}
                            onChange={(e) => setPeriodFrequency(e.target.value)}
                            disabled={busy}
                        >
                            <option value="monthly">Mensual</option>
                            <option value="biweekly">Quincenal</option>
                        </select>
                    </FormField>
                </div>

                <button
                    type="submit"
                    className="btn btnPrimary btnFull u-mt-6"
                    disabled={busy || !hhName.trim()}
                >
                    {busy ? <><span className="spinIcon" aria-hidden>⟳</span> Creando hogar…</> : 'Crear hogar →'}
                </button>
            </form>
        </section>
    )
}
