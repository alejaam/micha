import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
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
            setError('Passwords do not match.')
            return
        }
        setBusy(true)
        setError('')
        try {
            await register({ email: email.trim(), password })
            // Auto-login with the same credentials — no need to type them again
            await login({ email: email.trim(), password })
            navigate('/', { replace: true })
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

                <FormField label="Confirm password" htmlFor="regConfirmPassword">
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
                        <p className="formHint formHintError">Passwords do not match</p>
                    )}
                </FormField>

                <button type="submit" className="btn btnPrimary btnFull" disabled={!canSubmit}>
                    {busy ? <><span className="spinIcon" aria-hidden /> Creating account…</> : 'Create account'}
                </button>
            </form>
        </section>
    )
}
