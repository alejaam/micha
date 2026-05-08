import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createHousehold, createMember } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { useFormField } from '../hooks/useFormField'
import {
    AuthCard,
    AuthHeader,
    AuthFormField,
    AuthInput,
    AuthButton,
    AuthBanner,
} from '../ui/auth'
import { dollarsToCents } from '../utils'

const CURRENCIES = [
    { code: 'MXN', label: 'MXN — Peso mexicano' },
    { code: 'USD', label: 'USD — Dólar estadounidense' },
    { code: 'EUR', label: 'EUR — Euro' },
    { code: 'COP', label: 'COP — Peso colombiano' },
    { code: 'ARS', label: 'ARS — Peso argentino' },
    { code: 'CLP', label: 'CLP — Peso chileno' },
    { code: 'PEN', label: 'PEN — Sol peruano' },
    { code: 'BRL', label: 'BRL — Real brasileño' },
]

const SETTLEMENT_HINTS = {
    equal: 'Todos los miembros pagan la misma cantidad sin importar sus ingresos.',
    proportional:
        'Quienes ganan más contribuyen con una mayor parte de los gastos.',
}

export function OnboardingHouseholdPage() {
    const { user, handleProtectedError } = useAuth()
    const { setHouseholdId, loadHouseholds } = useAppShell()
    const navigate = useNavigate()

    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    const nombreHogar = useFormField('', (v) =>
        v.trim() ? null : 'El nombre es obligatorio.',
    )
    const modoSplit = useFormField('equal', () => null)
    const moneda = useFormField('MXN', () => null)
    const diaCierre = useFormField('15', (v) => {
        if (!v.trim()) return null
        const n = Number(v)
        if (!Number.isInteger(n) || n < 1 || n > 31)
            return 'Ingresa un día entre 1 y 31.'
        return null
    })
    const frecuencia = useFormField('monthly', () => null)
    const nombreMiembro = useFormField('', (v) =>
        v.trim() ? null : 'Tu nombre es obligatorio.',
    )
    const salario = useFormField('', (v) => {
        if (!v.trim()) return null
        const n = Number(v)
        if (isNaN(n) || n < 0) return 'Ingresa un valor válido.'
        return null
    })

    const canSubmit =
        nombreHogar.value.trim() !== '' &&
        nombreMiembro.value.trim() !== '' &&
        !busy

    async function handleSubmit(e) {
        e.preventDefault()
        nombreHogar.setTouched(true)
        nombreMiembro.setTouched(true)
        diaCierre.setTouched(true)
        if (!canSubmit) return

        setBusy(true)
        setError('')

        try {
            const hhOut = await createHousehold({
                name: nombreHogar.value.trim(),
                settlementMode: modoSplit.value,
                currency: moneda.value,
                closingDay: Number(diaCierre.value),
                periodFrequency: frecuencia.value,
            })

            const createdHouseholdId = hhOut?.household_id ?? hhOut?.id ?? ''
            if (!createdHouseholdId) {
                throw new Error(
                    'El hogar se creó pero no se devolvió el identificador.',
                )
            }

            setHouseholdId(createdHouseholdId)

            const salaryCents = dollarsToCents(salario.value) || 0
            await createMember({
                householdId: createdHouseholdId,
                name: nombreMiembro.value.trim(),
                email: user?.email || '',
                monthlySalaryCents: salaryCents,
            })

            await loadHouseholds()
            navigate('/onboarding/cards', { replace: true })
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <AuthCard>
                <AuthHeader
                    eyebrow="Paso 1 de 3"
                    title="Crear tu hogar"
                    subtitle="Un hogar agrupa todos los gastos y miembros compartidos."
                />

                {error ? <AuthBanner type="error">{error}</AuthBanner> : null}

                <form onSubmit={handleSubmit} noValidate aria-label="Crear hogar">
                    <AuthFormField
                        label="Nombre del hogar"
                        htmlFor="hhNombre"
                        error={nombreHogar.error}
                    >
                        <AuthInput
                            id="hhNombre"
                            placeholder="Ej. Casa Familia"
                            value={nombreHogar.value}
                            onChange={(e) => nombreHogar.setValue(e.target.value)}
                            onBlur={nombreHogar.onBlur}
                            disabled={busy}
                            hasError={!!nombreHogar.error}
                            autoFocus
                        />
                    </AuthFormField>

                    <AuthFormField label="Modo de reparto" htmlFor="hhModo">
                        <select
                            id="hhModo"
                            className="pd-input"
                            value={modoSplit.value}
                            onChange={(e) => modoSplit.setValue(e.target.value)}
                            disabled={busy}
                        >
                            <option value="equal">
                                Dividir en partes iguales
                            </option>
                            <option value="proportional">
                                Proporcional al salario
                            </option>
                        </select>
                        <p className="pd-hint">
                            {SETTLEMENT_HINTS[modoSplit.value]}
                        </p>
                    </AuthFormField>

                    <AuthFormField label="Moneda" htmlFor="hhMoneda">
                        <select
                            id="hhMoneda"
                            className="pd-input"
                            value={moneda.value}
                            onChange={(e) => moneda.setValue(e.target.value)}
                            disabled={busy}
                        >
                            {CURRENCIES.map((c) => (
                                <option key={c.code} value={c.code}>
                                    {c.label}
                                </option>
                            ))}
                        </select>
                    </AuthFormField>

                    <AuthFormField
                        label="Día de cierre"
                        htmlFor="hhCierre"
                        error={diaCierre.error}
                    >
                        <AuthInput
                            id="hhCierre"
                            type="number"
                            min="1"
                            max="31"
                            placeholder="15"
                            value={diaCierre.value}
                            onChange={(e) => diaCierre.setValue(e.target.value)}
                            onBlur={diaCierre.onBlur}
                            disabled={busy}
                            hasError={!!diaCierre.error}
                        />
                        <p className="pd-hint">
                            Día del mes en que se cierra el periodo.
                        </p>
                    </AuthFormField>

                    <AuthFormField
                        label="Frecuencia del periodo"
                        htmlFor="hhFrecuencia"
                    >
                        <select
                            id="hhFrecuencia"
                            className="pd-input"
                            value={frecuencia.value}
                            onChange={(e) =>
                                frecuencia.setValue(e.target.value)
                            }
                            disabled={busy}
                        >
                            <option value="monthly">Mensual</option>
                            <option value="biweekly">Quincenal</option>
                        </select>
                    </AuthFormField>

                    <hr className="pd-divider" />

                    <p className="pd-note">
                        Vas a ser agregado como el primer miembro. Tu correo (
                        {user?.email}) se vincula automáticamente.
                    </p>

                    <AuthFormField
                        label="Tu nombre"
                        htmlFor="memNombre"
                        error={nombreMiembro.error}
                    >
                        <AuthInput
                            id="memNombre"
                            placeholder="Ej. Alex"
                            value={nombreMiembro.value}
                            onChange={(e) =>
                                nombreMiembro.setValue(e.target.value)
                            }
                            onBlur={nombreMiembro.onBlur}
                            disabled={busy}
                            hasError={!!nombreMiembro.error}
                        />
                    </AuthFormField>

                    <AuthFormField
                        label="Salario mensual (opcional)"
                        htmlFor="memSalario"
                        error={salario.error}
                    >
                        <div className="pd-inputGroup">
                            <span className="pd-inputPrefix" aria-hidden>
                                $
                            </span>
                            <AuthInput
                                id="memSalario"
                                type="number"
                                min="0"
                                step="0.01"
                                placeholder="Ej. 30000"
                                value={salario.value}
                                onChange={(e) =>
                                    salario.setValue(e.target.value)
                                }
                                onBlur={salario.onBlur}
                                disabled={busy}
                                hasError={!!salario.error}
                                className="pd-inputWithPrefix"
                            />
                        </div>
                        <p className="pd-hint">
                            Se usa para calcular repartos proporcionales. Podés
                            actualizarlo después.
                        </p>
                    </AuthFormField>

                    <AuthButton
                        type="submit"
                        fullWidth
                        disabled={!canSubmit}
                        busy={busy}
                    >
                        {busy ? 'Guardando…' : 'Continuar'}
                    </AuthButton>
                </form>
            </AuthCard>
    )
}
