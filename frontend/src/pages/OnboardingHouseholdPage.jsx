import { useEffect, useMemo, useState } from 'react'
import { motion } from 'framer-motion'
import { useLocation, useNavigate } from 'react-router-dom'
import { createHousehold, updateHousehold } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'
import { dollarsToCents, sanitizeAmountInput } from '../utils'

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
    const { householdId, selectedHousehold, setHouseholdId, loadHouseholds } = useAppShell()
    const navigate = useNavigate()
    const location = useLocation()

    const isOnboarding = location.pathname.startsWith('/onboarding/')

    const initial = useMemo(() => {
        if (isOnboarding) {
            return {
                name: '',
                settlementMode: 'equal',
                currency: 'MXN',
                closingDay: 15,
                periodFrequency: 'biweekly',
                salary: '',
            }
        }
        return {
            name: selectedHousehold?.name ?? '',
            settlementMode: selectedHousehold?.settlement_mode ?? 'equal',
            currency: selectedHousehold?.currency ?? 'MXN',
            closingDay: selectedHousehold?.closing_day ?? 15,
            periodFrequency: selectedHousehold?.period_frequency ?? 'biweekly',
            salary: '',
        }
    }, [isOnboarding, selectedHousehold])

    const [hhName, setHhName] = useState(initial.name)
    const [settlementMode, setSettlementMode] = useState(initial.settlementMode)
    const [currency, setCurrency] = useState(initial.currency)
    const [closingDay, setClosingDay] = useState(initial.closingDay)
    const [periodFrequency, setPeriodFrequency] = useState(initial.periodFrequency)
    const [salary, setSalary] = useState(initial.salary)

    useEffect(() => {
        if (isOnboarding) return
        setHhName(initial.name)
        setSettlementMode(initial.settlementMode)
        setCurrency(initial.currency)
        setClosingDay(initial.closingDay)
        setPeriodFrequency(initial.periodFrequency)
    }, [isOnboarding, initial])

    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    async function handleSubmit(e) {
        e.preventDefault()
        if (!hhName.trim()) return

        setBusy(true)
        setError('')

        try {
            if (!isOnboarding) {
                if (!householdId) {
                    throw new Error('No hay hogar seleccionado para editar')
                }

                await updateHousehold({
                    householdId,
                    name: hhName.trim(),
                    settlementMode,
                    currency,
                    closingDay: Number(closingDay),
                    periodFrequency,
                })

                await loadHouseholds()
                navigate('/rules')
                return
            }

            // Onboarding: create household with owner salary
            const salaryCents = dollarsToCents(salary) || 0
            const hhOut = await createHousehold({
                name: hhName.trim(),
                settlementMode,
                currency,
                closingDay: Number(closingDay),
                periodFrequency,
                ownerSalaryCents: salaryCents,
            })

            const createdHouseholdId = hhOut?.household_id ?? hhOut?.id ?? ''
            if (!createdHouseholdId) {
                throw new Error('El hogar fue creado pero no se devolvió el ID')
            }

            setHouseholdId(createdHouseholdId)
            await loadHouseholds()

            // Redirect to cards onboarding (skipping the separate member step)
            navigate('/onboarding/cards', { replace: true })

        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return isOnboarding ? (
        <section className="card onboardingCard" aria-label="Crea tu hogar">
            <div className="onboardingHeader">
                <p className="authEyebrow">Primeros pasos</p>
                <h2 className="authTitle">Configura tu hogar</h2>
                <p className="authMeta">
                    Un hogar agrupa todos los gastos compartidos y los miembros.
                </p>
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

                    {isOnboarding && (
                        <FormField label="Tu salario mensual (opcional)" htmlFor="hhSalary">
                            <div className="inputWrap">
                                <span className="inputPrefix" aria-hidden>$</span>
                                <input
                                    id="hhSalary"
                                    className="input inputWithPrefix"
                                    type="text"
                                    inputMode="decimal"
                                    placeholder="0.00"
                                    value={salary}
                                    onChange={(e) => setSalary(sanitizeAmountInput(e.target.value))}
                                    disabled={busy}
                                />
                            </div>
                            <p className="formHint">Tu salario mensual bruto. Se usará para calcular la distribución proporcional de gastos.</p>
                        </FormField>
                    )}
                </div>

                <div className="u-flex u-gap-4 u-mt-6">
                    <button
                        type="submit"
                        className="btn btnPrimary u-flex-1 btnFull"
                        disabled={busy || !hhName.trim()}
                    >
                        {busy
                            ? <><span className="spinIcon" aria-hidden>⟳</span> Creando hogar…</>
                            : 'Crear hogar →'}
                    </button>
                </div>
            </form>
        </section>
    ) : (
        <motion.div
            className="pageGrid"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.2 }}
        >
            <div className="dashboardCol">
                <section className="card" aria-label="Editar hogar">
                    <div className="listHeader">
                        <h2 className="listTitle">Editar hogar</h2>
                    </div>
                    <p className="u-text-sm u-text-dim u-mb-3">
                        Actualiza el nombre, moneda y configuración del periodo de tu hogar.
                    </p>

                    {error ? <Banner type="error" floating onDismiss={() => setError('')}>{error}</Banner> : null}

                    <form className="formStack" onSubmit={handleSubmit}>
                        <div className="formSection">
                            <h3 className="sectionTitle">Detalles del hogar</h3>
                            <FormField label="Nombre del hogar" htmlFor="hhNameSettings">
                                <input
                                    id="hhNameSettings"
                                    className="input"
                                    placeholder="Ej. Casa Familia"
                                    value={hhName}
                                    onChange={(e) => setHhName(e.target.value)}
                                    disabled={busy}
                                    autoFocus
                                />
                            </FormField>
                            <FormField label="Modo de liquidación" htmlFor="hhModeSettings">
                                <select
                                    id="hhModeSettings"
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
                            <FormField label="Moneda" htmlFor="hhCurrencySettings">
                                <select
                                    id="hhCurrencySettings"
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

                            <FormField label="Día de cierre" htmlFor="hhClosingDaySettings">
                                <input
                                    id="hhClosingDaySettings"
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

                            <FormField label="Frecuencia del periodo" htmlFor="hhFrequencySettings">
                                <select
                                    id="hhFrequencySettings"
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

                        <div className="u-flex u-gap-4 u-mt-6">
                            <button
                                type="button"
                                className="btn u-flex-1"
                                onClick={() => navigate('/rules')}
                                disabled={busy}
                            >
                                ← Volver a ajustes
                            </button>
                            <button
                                type="submit"
                                className="btn btnPrimary u-flex-1"
                                disabled={busy || !hhName.trim()}
                            >
                                {busy
                                    ? <><span className="spinIcon" aria-hidden>⟳</span> Guardando…</>
                                    : 'Guardar cambios'}
                            </button>
                        </div>
                    </form>
                </section>
            </div>
        </motion.div>
    )
}
