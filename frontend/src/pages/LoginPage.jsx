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
        <section className="authCard card" aria-label="Sign in">
            <div className="authHeader">
                <p className="authEyebrow">Welcome to micha</p>
                <h1 className="authTitle">Sign in to your household</h1>
                <p className="authMeta">Use your registered email and password to continue.</p>
            </div>

            <div className="authSwitch">
                <span className="btn btnPrimary btnSm">Sign in</span>
                <Link to="/register" className="btn btnGhost btnSm">Create account</Link>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit} noValidate>
                <FormField label="Email" htmlFor="loginEmail">
                    <input
                        id="loginEmail"
                        className="input"
                        type="email"
                        autoComplete="email"
                        placeholder="you@example.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        disabled={busy}
                    />
                </FormField>

                <FormField label="Password" htmlFor="loginPassword">
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
                    {busy ? <><span className="spinIcon" aria-hidden>⟳</span> Signing in…</> : 'Sign in'}
                </button>
            </form>
        </section>
    )
}
