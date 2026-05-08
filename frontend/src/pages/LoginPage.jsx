import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useFormField } from '../hooks/useFormField'
import { AuthCard, AuthHeader, AuthFormField, AuthInput, AuthButton, AuthBanner } from '../ui/auth'

export function LoginPage() {
    const { login } = useAuth()
    const navigate = useNavigate()
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    const email = useFormField('', (v) => (v.trim() ? null : 'Ingresa un correo válido.'))
    const password = useFormField('', (v) => (v ? null : 'Ingresa tu contraseña.'))

    const canSubmit = email.value.trim() !== '' && password.value.trim() !== '' && !busy

    async function handleSubmit(e) {
        e.preventDefault()
        email.setTouched(true)
        password.setTouched(true)
        if (!canSubmit) return

        setBusy(true)
        setError('')
        try {
            await login({ email: email.value.trim(), password: password.value })
            navigate('/', { replace: true })
        } catch (err) {
            setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <AuthCard>
            <AuthHeader
                eyebrow="Bienvenido a micha"
                title="Iniciar sesión"
                subtitle="Usa tu correo y contraseña registrados para continuar."
            />

            <div style={{ display: 'flex', gap: 8, marginBottom: 24 }}>
                <span className="pd-btn pd-btnPrimary" style={{ flex: 1, textAlign: 'center' }}>
                    Iniciar sesión
                </span>
                <Link
                    to="/register"
                    className="pd-btn pd-btnGhost"
                    style={{ flex: 1, textAlign: 'center', textDecoration: 'none' }}
                >
                    Crear cuenta
                </Link>
            </div>

            {error ? <AuthBanner type="error">{error}</AuthBanner> : null}

            <form onSubmit={handleSubmit} noValidate aria-label="Iniciar sesión">
                <AuthFormField label="Correo electrónico" htmlFor="loginEmail" error={email.error}>
                    <AuthInput
                        id="loginEmail"
                        type="email"
                        autoComplete="email"
                        placeholder="tucorreo@ejemplo.com"
                        value={email.value}
                        onChange={(e) => email.setValue(e.target.value)}
                        onBlur={email.onBlur}
                        disabled={busy}
                        hasError={!!email.error}
                    />
                </AuthFormField>

                <AuthFormField label="Contraseña" htmlFor="loginPassword" error={password.error}>
                    <AuthInput
                        id="loginPassword"
                        type="password"
                        autoComplete="current-password"
                        placeholder="••••••••"
                        value={password.value}
                        onChange={(e) => password.setValue(e.target.value)}
                        onBlur={password.onBlur}
                        disabled={busy}
                        hasError={!!password.error}
                    />
                </AuthFormField>

                <AuthButton type="submit" fullWidth disabled={!canSubmit} busy={busy}>
                    {busy ? 'Iniciando sesión…' : 'Iniciar sesión'}
                </AuthButton>
            </form>
        </AuthCard>
    )
}
