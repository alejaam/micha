import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createHousehold, createMember } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'
import { dollarsToCents } from '../utils'

const CURRENCIES = [
    { code: 'MXN', label: '🇲🇽 MXN — Mexican Peso' },
    { code: 'USD', label: '🇺🇸 USD — US Dollar' },
    { code: 'EUR', label: '🇪🇺 EUR — Euro' },
    { code: 'COP', label: '🇨🇴 COP — Colombian Peso' },
    { code: 'ARS', label: '🇦🇷 ARS — Argentine Peso' },
    { code: 'CLP', label: '🇨🇱 CLP — Chilean Peso' },
    { code: 'PEN', label: '🇵🇪 PEN — Peruvian Sol' },
    { code: 'BRL', label: '🇧🇷 BRL — Brazilian Real' },
]

const SETTLEMENT_HINTS = {
    equal: 'Cada miembro paga la misma parte, sin importar ingresos.',
    proportional: 'Los miembros que ganan más contribuyen con una mayor parte de los gastos.',
}

export function OnboardingHouseholdPage() {
    const { user, handleProtectedError } = useAuth()
    const { setHouseholdId, loadHouseholds } = useAppShell()
    const navigate = useNavigate()

    // Household state
    const [hhName, setHhName] = useState('')
    const [settlementMode, setSettlementMode] = useState('equal')
    const [currency, setCurrency] = useState('MXN')
    const [closingDay, setClosingDay] = useState(15)
    const [periodFrequency, setPeriodFrequency] = useState('monthly')

    // Member state
    const [memberName, setMemberName] = useState('')
    const [salaryDollars, setSalaryDollars] = useState('')

    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    async function handleSubmit(e) {
        e.preventDefault()
        if (!hhName.trim() || !memberName.trim()) return
        
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
                throw new Error('household created but id was not returned')
            }

            // Keep the new ID locally
            setHouseholdId(createdHouseholdId)

            // 2. Auto-create the creator as the first member
            const salaryCents = dollarsToCents(salaryDollars) || 0
            await createMember({
                householdId: createdHouseholdId,
                name: memberName.trim(),
                email: user?.email || '',
                monthlySalaryCents: salaryCents,
            })

            // 3. Refresh households list now that there's a member linked to the user
            await loadHouseholds()

            // 4. Continue onboarding with cards setup
            navigate('/onboarding/cards', { replace: true })
            
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

            {error ? <Banner type="error">{error}</Banner> : null}

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

                <div className="formSection u-mt-4">
                    <h3 className="sectionTitle">Tu perfil</h3>
                    <p className="u-text-sm u-text-dim u-mb-2">
                        Serás añadido como el primer miembro. Tu correo ({user?.email}) se vincula automáticamente.
                    </p>
                    <FormField label="Tu nombre" htmlFor="memName">
                        <input
                            id="memName"
                            className="input"
                            placeholder="Ej. Alex"
                            value={memberName}
                            onChange={(e) => setMemberName(e.target.value)}
                            disabled={busy}
                        />
                    </FormField>
                    <FormField label="Salario mensual (opcional)" htmlFor="memSalary">
                        <div className="inputWrap">
                            <span className="inputPrefix" aria-hidden>$</span>
                            <input
                                id="memSalary"
                                className="input inputWithPrefix"
                                type="number"
                                min="0"
                                step="0.01"
                                placeholder="Ej. 30000"
                                value={salaryDollars}
                                onChange={(e) => setSalaryDollars(e.target.value)}
                                disabled={busy}
                            />
                        </div>
                        <p className="formHint">Se usa para calcular la división proporcional. Puedes actualizarlo después.</p>
                    </FormField>
                </div>

                <button
                    type="submit"
                    className="btn btnPrimary btnFull u-mt-6"
                    disabled={busy || !hhName.trim() || !memberName.trim()}
                >
                    {busy ? <><span className="spinIcon" aria-hidden>⟳</span> Creando…</> : 'Finalizar configuración →'}
                </button>
            </form>
        </section>
    )
}
