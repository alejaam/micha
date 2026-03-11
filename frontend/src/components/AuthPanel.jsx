import { useMemo, useState } from 'react'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

/**
 * AuthPanel renders login/register form with shared UX behavior.
 */
export function AuthPanel({ mode, onModeChange, onLogin, onRegister, isSubmitting, error, message }) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const isLogin = mode === 'login'
  const canSubmit = useMemo(
    () => email.trim() !== '' && password.trim() !== '' && !isSubmitting,
    [email, password, isSubmitting],
  )

  async function handleSubmit(event) {
    event.preventDefault()

    const payload = {
      email: email.trim(),
      password,
    }

    if (isLogin) {
      await onLogin(payload)
      return
    }

    await onRegister(payload)
  }

  return (
    <section className="authCard card" aria-label="Authentication">
      <div className="authHeader">
        <p className="authEyebrow">Welcome to micha</p>
        <h1 className="authTitle">{isLogin ? 'Sign in to your household' : 'Create your account'}</h1>
        <p className="authMeta">
          {isLogin
            ? 'Use your registered email and password to continue.'
            : 'Create credentials to start tracking shared expenses.'}
        </p>
      </div>

      <div className="authSwitch" role="tablist" aria-label="Authentication mode">
        <button
          type="button"
          className={isLogin ? 'btn btnPrimary btnSm' : 'btn btnGhost btnSm'}
          onClick={() => onModeChange('login')}
          aria-selected={isLogin}
          role="tab"
          disabled={isSubmitting}
        >
          Sign in
        </button>
        <button
          type="button"
          className={!isLogin ? 'btn btnPrimary btnSm' : 'btn btnGhost btnSm'}
          onClick={() => onModeChange('register')}
          aria-selected={!isLogin}
          role="tab"
          disabled={isSubmitting}
        >
          Create account
        </button>
      </div>

      {error ? <Banner type="error">{error}</Banner> : null}
      {message ? <Banner type="ok">{message}</Banner> : null}

      <form className="formStack" onSubmit={handleSubmit} noValidate>
        <FormField label="Email" htmlFor="authEmail">
          <input
            id="authEmail"
            className="input"
            type="email"
            autoComplete="email"
            placeholder="you@example.com"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            disabled={isSubmitting}
          />
        </FormField>

        <FormField label="Password" htmlFor="authPassword">
          <input
            id="authPassword"
            className="input"
            type="password"
            autoComplete={isLogin ? 'current-password' : 'new-password'}
            placeholder="••••••••"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            disabled={isSubmitting}
          />
        </FormField>

        <button type="submit" className="btn btnPrimary btnFull" disabled={!canSubmit}>
          {isSubmitting
            ? <><span className="spinIcon" aria-hidden>⟳</span> Working...</>
            : isLogin ? 'Sign in' : 'Create account'}
        </button>
      </form>
    </section>
  )
}
