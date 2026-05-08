import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useFormField } from '../hooks/useFormField'
import { AuthCard, AuthHeader, AuthFormField, AuthInput, AuthButton, AuthBanner } from '../ui/auth'

export function RegisterPage() {
    const { register, login } = useAuth()
    const navigate = useNavigate()
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    const email = useFormField('', (v) => (v.trim() ? null : 'Ingresa un correo válido.'))
    const password = useFormField('', (v) => {
        if (!v) return 'Ingresa tu contraseña.'
        if (v.length < 6) return 'Mínimo 6 caracteres.'
        return null
    })

    const confirmPassword = useFormField('', (v) => {
        if (!v) return 'Confirma tu contraseña.'
        if (v !== password.value) return 'Las contraseñas no coinciden.'
        return null
    })

    const passwordsMatch = password.value === confirmPassword.value
    const canSubmit =
        email.value.trim() !== '' &&
        password.value.trim() !== '' &&
        confirmPassword.value.trim() !== '' &&
        passwordsMatch &&
        !busy

    async function handleSubmit(e) {
        e.preventDefault()
        email.setTouched(true)
        password.setTouched(true)
        confirmPassword.setTouched(true)
        if (!canSubmit) return

        setBusy(true)
        setError('')
        try {
            await register({ email: email.value.trim(), password: password.value })
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
                title="Crear tu cuenta"
                subtitle="Creá tus credenciales para empezar a gestionar gastos compartidos."
            />

            <div style={{ display: 'flex', gap: 8, marginBottom: 24 }}>
                <Link
                    to="/login"
                    className="pd-btn pd-btnGhost"
                    style={{ flex: 1, textAlign: 'center', textDecoration: 'none' }}
                >
                    Iniciar sesión
                </Link>
                <span className="pd-btn pd-btnPrimary" style={{ flex: 1, textAlign: 'center' }}>
                    Crear cuenta
                </span>
            </div>

            {error ? <AuthBanner type="error">{error}</AuthBanner> : null}

            <form onSubmit={handleSubmit} noValidate aria-label="Crear cuenta">
                <AuthFormField label="Correo electrónico" htmlFor="regEmail" error={email.error}>
                    <AuthInput
                        id="regEmail"
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

                <AuthFormField label="Contraseña" htmlFor="regPassword" error={password.error}>
                    <AuthInput
                        id="regPassword"
                        type="password"
                        autoComplete="new-password"
                        placeholder="••••••••"
                        value={password.value}
                        onChange={(e) => password.setValue(e.target.value)}
                        onBlur={password.onBlur}
                        disabled={busy}
                        hasError={!!password.error}
                    />
                </AuthFormField>

                <AuthFormField label="Confirmar contraseña" htmlFor="regConfirmPassword" error={confirmPassword.error}>
                    <AuthInput
                        id="regConfirmPassword"
                        type="password"
                        autoComplete="new-password"
                        placeholder="••••••••"
                        value={confirmPassword.value}
                        onChange={(e) => confirmPassword.setValue(e.target.value)}
                        onBlur={confirmPassword.onBlur}
                        disabled={busy}
                        hasError={!!confirmPassword.error}
                    />
                </AuthFormField>

                <AuthButton type="submit" fullWidth disabled={!canSubmit} busy={busy}>
                    {busy ? 'Creando cuenta…' : 'Crear cuenta'}
                </AuthButton>
            </form>
        </AuthCard>
    )
}
