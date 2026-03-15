import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createHousehold } from '../api'
import { useAuth } from '../context/AuthContext'
import { useAppShell } from '../context/AppShellContext'
import { FormField } from '../ui/FormField'
import { Banner } from '../ui/Banner'

export function OnboardingHouseholdPage() {
    const { handleProtectedError } = useAuth()
    const { loadHouseholds, setHouseholdId } = useAppShell()
    const navigate = useNavigate()
    const [name, setName] = useState('')
    const [settlementMode, setSettlementMode] = useState('equal')
    const [currency, setCurrency] = useState('MXN')
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    async function handleSubmit(e) {
        e.preventDefault()
        if (!name.trim()) return
        setBusy(true)
        setError('')
        try {
            const out = await createHousehold({
                name: name.trim(),
                settlementMode,
                currency: currency.trim().toUpperCase() || 'MXN',
            })
            await loadHouseholds()
            if (out?.household_id) setHouseholdId(out.household_id)
            navigate('/onboarding/member', { replace: true })
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <section className="card onboardingCard" aria-label="Create first household">
            <div className="onboardingHeader">
                <p className="authEyebrow">Step 1 of 2</p>
                <h2 className="authTitle">Create your household</h2>
                <p className="authMeta">A household groups all shared expenses and members.</p>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit}>
                <FormField label="Name" htmlFor="hhName">
                    <input
                        id="hhName"
                        className="input"
                        placeholder="e.g. Casa Familia"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        disabled={busy}
                    />
                </FormField>
                <FormField label="Settlement mode" htmlFor="hhMode">
                    <select
                        id="hhMode"
                        className="input"
                        value={settlementMode}
                        onChange={(e) => setSettlementMode(e.target.value)}
                        disabled={busy}
                    >
                        <option value="equal">Equal split</option>
                        <option value="proportional">Proportional to salary</option>
                    </select>
                </FormField>
                <FormField label="Currency" htmlFor="hhCurrency">
                    <input
                        id="hhCurrency"
                        className="input"
                        placeholder="MXN"
                        value={currency}
                        onChange={(e) => setCurrency(e.target.value)}
                        disabled={busy}
                    />
                </FormField>
                <button
                    type="submit"
                    className="btn btnPrimary btnFull"
                    disabled={busy || !name.trim()}
                >
                    {busy ? <><span className="spinIcon" aria-hidden>⟳</span> Creating…</> : 'Create household →'}
                </button>
            </form>
        </section>
    )
}
