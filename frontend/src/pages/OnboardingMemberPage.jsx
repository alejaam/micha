import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createMember } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'
import { dollarsToCents } from '../utils'

export function OnboardingMemberPage() {
    const { handleProtectedError } = useAuth()
    const { householdId, loadHouseholds } = useAppShell()
    const navigate = useNavigate()
    
    const [name, setName] = useState('')
    const [email, setEmail] = useState('')
    const [salaryDollars, setSalaryDollars] = useState('')
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    async function handleSubmit(e) {
        e.preventDefault()
        if (!householdId || !name.trim() || !email.trim()) return
        setBusy(true)
        setError('')
        try {
            const salaryCents = dollarsToCents(salaryDollars) || 0
            await createMember({
                householdId: householdId,
                name: name.trim(),
                email: email.trim(),
                monthlySalaryCents: salaryCents,
            })
            // Refresh households just in case this member triggers something, 
            // though usually it only affects member list
            await loadHouseholds()
            navigate('/', { replace: true })
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    if (!householdId) {
        return (
            <div className="card">
                <Banner type="error">No household selected. Go back to dashboard.</Banner>
                <button className="btn mt-4" onClick={() => navigate('/')}>Back</button>
            </div>
        )
    }

    return (
        <section className="card" aria-label="Add a member">
            <div className="listHeader mb-6">
                <div>
                    <h2 className="listTitle">Add a member</h2>
                    <p className="text-sm text-dim mt-1">
                        Add someone to your household to track their expenses.
                    </p>
                </div>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit}>
                <FormField label="Name" htmlFor="memName">
                    <input
                        id="memName"
                        className="input"
                        placeholder="e.g. Maria"
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
                        placeholder="maria@example.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        disabled={busy}
                    />
                </FormField>
                <FormField label="Monthly salary (optional)" htmlFor="memSalary">
                    <div className="inputWrap">
                        <span className="inputPrefix" aria-hidden>$</span>
                        <input
                            id="memSalary"
                            className="input inputWithPrefix"
                            type="number"
                            min="0"
                            step="0.01"
                            placeholder="e.g., 5000"
                            value={salaryDollars}
                            onChange={(e) => setSalaryDollars(e.target.value)}
                            disabled={busy}
                        />
                    </div>
                </FormField>
                <div className="flex gap-4 mt-6">
                    <button
                        type="button"
                        className="btn flex-1"
                        onClick={() => navigate('/')}
                        disabled={busy}
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        className="btn btnPrimary flex-1"
                        disabled={busy || !name.trim() || !email.trim()}
                    >
                        {busy ? <><span className="spinIcon" aria-hidden /> Adding…</> : 'Add member'}
                    </button>
                </div>
            </form>
        </section>
    )
}
