import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

export function LoginPage() {
    const { login } = useAuth()
    const navigate = useNavigate()
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    const canSubmit = email.trim() !== '' && password.trim() !== '' && !busy

    async function handleSubmit(e) {
        e.preventDefault()
        setBusy(true)
        setError('')
        try {
            await login({ email: email.trim(), password })
            navigate('/', { replace: true })
        } catch (err) {
            setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <section className="authCard card" aria-label="Iniciar sesión">
            <div className="authHeader">
                <p className="authEyebrow">Bienvenido a micha</p>
                <h1 className="authTitle">Iniciar sesión en tu hogar</h1>
                <p className="authMeta">Usa tu correo y contraseña registrados para continuar.</p>
            </div>

            <div className="authSwitch">
                <span className="btn btnPrimary btnSm">Iniciar sesión</span>
                <Link to="/register" className="btn btnGhost btnSm">Crear cuenta</Link>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit} noValidate>
                <FormField label="Correo electrónico" htmlFor="loginEmail">
                    <input
                        id="loginEmail"
                        className="input"
                        type="email"
                        autoComplete="email"
                        placeholder="tu@ejemplo.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        disabled={busy}
                    />
                </FormField>

                <FormField label="Contraseña" htmlFor="loginPassword">
                    <input
                        id="loginPassword"
                        className="input"
                        type="password"
                        autoComplete="current-password"
                        placeholder="••••••••"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        disabled={busy}
                    />
                </FormField>

                <button type="submit" className="btn btnPrimary btnFull" disabled={!canSubmit}>
                    {busy ? <><span className="spinIcon" aria-hidden>⟳</span> Iniciando sesión…</> : 'Iniciar sesión'}
                </button>
            </form>
        </section>
    )
}
