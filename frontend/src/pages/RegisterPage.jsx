import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

export function RegisterPage() {
    const { register, login } = useAuth()
    const navigate = useNavigate()
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [confirmPassword, setConfirmPassword] = useState('')
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    const passwordsMatch = password === confirmPassword
    const canSubmit =
        email.trim() !== '' &&
        password.trim() !== '' &&
        confirmPassword.trim() !== '' &&
        passwordsMatch &&
        !busy

    async function handleSubmit(e) {
        e.preventDefault()
        if (!passwordsMatch) {
            setError('Las contraseñas no coinciden.')
            return
        }
        setBusy(true)
        setError('')
        try {
            await register({ email: email.trim(), password })
            // Auto-login with the same credentials — no need to type them again
            await login({ email: email.trim(), password })
            navigate('/onboarding/household', { replace: true })
        } catch (err) {
            setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <section className="authCard card" aria-label="Crear cuenta">
            <div className="authHeader">
                <p className="authEyebrow">Bienvenido a micha</p>
                <h1 className="authTitle">Crear tu cuenta</h1>
                <p className="authMeta">Crea tus credenciales para empezar a registrar gastos compartidos.</p>
            </div>

            <div className="authSwitch">
                <Link to="/login" className="btn btnGhost btnSm">Iniciar sesión</Link>
                <span className="btn btnPrimary btnSm">Crear cuenta</span>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit} noValidate>
                <FormField label="Correo electrónico" htmlFor="regEmail">
                    <input
                        id="regEmail"
                        className="input"
                        type="email"
                        autoComplete="email"
                        placeholder="tu@ejemplo.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        disabled={busy}
                    />
                </FormField>

                <FormField label="Contraseña" htmlFor="regPassword">
                    <input
                        id="regPassword"
                        className="input"
                        type="password"
                        autoComplete="new-password"
                        placeholder="••••••••"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        disabled={busy}
                    />
                </FormField>

                <FormField label="Confirmar contraseña" htmlFor="regConfirmPassword">
                    <input
                        id="regConfirmPassword"
                        className="input"
                        type="password"
                        autoComplete="new-password"
                        placeholder="••••••••"
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        disabled={busy}
                    />
                    {confirmPassword && !passwordsMatch && (
                        <p className="formHint formHintError">Las contraseñas no coinciden</p>
                    )}
                </FormField>

                <button type="submit" className="btn btnPrimary btnFull" disabled={!canSubmit}>
                    {busy ? <><span className="spinIcon" aria-hidden>⟳</span> Creando cuenta…</> : 'Crear cuenta'}
                </button>
            </form>
        </section>
    )
}
