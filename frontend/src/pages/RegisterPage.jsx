import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

export function RegisterPage() {
    const { register } = useAuth()
    const navigate = useNavigate()
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')
    const [message, setMessage] = useState('')

    const canSubmit = email.trim() !== '' && password.trim() !== '' && !busy

    async function handleSubmit(e) {
        e.preventDefault()
        setBusy(true)
        setError('')
        setMessage('')
        try {
            await register({ email: email.trim(), password })
            setMessage('Account created. Sign in with your credentials.')
            setTimeout(() => navigate('/login', { replace: true }), 1200)
        } catch (err) {
            setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <section className="authCard card" aria-label="Create account">
            <div className="authHeader">
                <p className="authEyebrow">Welcome to micha</p>
                <h1 className="authTitle">Create your account</h1>
                <p className="authMeta">Create credentials to start tracking shared expenses.</p>
            </div>

            <div className="authSwitch">
                <Link to="/login" className="btn btnGhost btnSm">Sign in</Link>
                <span className="btn btnPrimary btnSm">Create account</span>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}
            {message ? <Banner type="ok">{message}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit} noValidate>
                <FormField label="Email" htmlFor="regEmail">
                    <input
                        id="regEmail"
                        className="input"
                        type="email"
                        autoComplete="email"
                        placeholder="you@example.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        disabled={busy}
                    />
                </FormField>

                <FormField label="Password" htmlFor="regPassword">
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

                <button type="submit" className="btn btnPrimary btnFull" disabled={!canSubmit}>
                    {busy ? <><span className="spinIcon" aria-hidden>⟳</span> Creating…</> : 'Create account'}
                </button>
            </form>
        </section>
    )
}
