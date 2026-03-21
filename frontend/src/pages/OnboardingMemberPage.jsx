import { useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { createMember } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

export function OnboardingMemberPage() {
    const { handleProtectedError } = useAuth()
    const { householdId, setHouseholdId } = useAppShell()
    const location = useLocation()
    const navigate = useNavigate()
    const [name, setName] = useState('')
    const [email, setEmail] = useState('')
    const [salary, setSalary] = useState('0')
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')
    const routeHouseholdId = location.state?.householdId ?? ''
    const activeHouseholdId = householdId || routeHouseholdId

    useEffect(() => {
        if (!householdId && routeHouseholdId) {
            setHouseholdId(routeHouseholdId)
        }
    }, [householdId, routeHouseholdId, setHouseholdId])

    useEffect(() => {
        if (!activeHouseholdId) {
            navigate('/onboarding/household', { replace: true })
        }
    }, [activeHouseholdId, navigate])

    async function handleSubmit(e) {
        e.preventDefault()
        if (!activeHouseholdId || !name.trim() || !email.trim()) return
        setBusy(true)
        setError('')
        try {
            await createMember({
                householdId: activeHouseholdId,
                name: name.trim(),
                email: email.trim(),
                monthlySalaryCents: Number(salary) || 0,
            })
            navigate('/', { replace: true })
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <section className="card onboardingCard" aria-label="Create first member">
            <div className="onboardingHeader">
                <p className="authEyebrow">Step 2 of 2</p>
                <h2 className="authTitle">Add yourself as a member</h2>
                <p className="authMeta">
                    Use the same email as your account so expenses are linked to you automatically.
                </p>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit}>
                <FormField label="Name" htmlFor="memName">
                    <input
                        id="memName"
                        className="input"
                        placeholder="e.g. Alex"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        disabled={busy}
                    />
                </FormField>
                <FormField label="Email" htmlFor="memEmail">
                    <input
                        id="memEmail"
                        className="input"
                        type="email"
                        placeholder="you@example.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        disabled={busy}
                    />
                </FormField>
                <FormField label="Monthly salary (cents, optional)" htmlFor="memSalary">
                    <input
                        id="memSalary"
                        className="input"
                        type="number"
                        min="0"
                        placeholder="0"
                        value={salary}
                        onChange={(e) => setSalary(e.target.value)}
                        disabled={busy}
                    />
                </FormField>
                <button
                    type="submit"
                    className="btn btnPrimary btnFull"
                    disabled={busy || !name.trim() || !email.trim()}
                >
                    {busy ? <><span className="spinIcon" aria-hidden>⟳</span> Creating…</> : 'Finish setup →'}
                </button>
            </form>
        </section>
    )
}
