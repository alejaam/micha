import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createHousehold, createMember } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'

export function OnboardingHouseholdPage() {
    const { user, handleProtectedError } = useAuth()
    const { setHouseholdId, loadHouseholds } = useAppShell()
    const navigate = useNavigate()

    // Household state
    const [hhName, setHhName] = useState('')
    const [settlementMode, setSettlementMode] = useState('equal')
    const [currency, setCurrency] = useState('MXN')

    // Member state
    const [memberName, setMemberName] = useState('')
    const [salary, setSalary] = useState('0')

    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    async function handleSubmit(e) {
        e.preventDefault()
        if (!hhName.trim() || !memberName.trim()) return
        
        setBusy(true)
        setError('')
        
        try {
            // 1. Create household
            const hhOut = await createHousehold({
                name: hhName.trim(),
                settlementMode,
                currency: currency.trim().toUpperCase() || 'MXN',
            })
            
            const createdHouseholdId = hhOut?.household_id ?? hhOut?.id ?? ''
            if (!createdHouseholdId) {
                throw new Error('household created but id was not returned')
            }

            // Keep the new ID locally
            setHouseholdId(createdHouseholdId)

            // 2. Auto-create the creator as the first member
            await createMember({
                householdId: createdHouseholdId,
                name: memberName.trim(),
                email: user?.email || '', // The backend will link the member to the user via this email
                monthlySalaryCents: Number(salary) || 0,
            })

            // 3. Refresh households list now that there's a member linked to the user
            await loadHouseholds()

            // 4. Finish onboarding
            navigate('/', { replace: true })
            
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <section className="card onboardingCard" aria-label="Create your household">
            <div className="onboardingHeader">
                <p className="authEyebrow">Getting started</p>
                <h2 className="authTitle">Set up your household</h2>
                <p className="authMeta">A household groups all shared expenses and members.</p>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit}>
                <div className="formSection">
                    <h3 className="sectionTitle">Household Details</h3>
                    <FormField label="Household name" htmlFor="hhName">
                        <input
                            id="hhName"
                            className="input"
                            placeholder="e.g. Casa Familia"
                            value={hhName}
                            onChange={(e) => setHhName(e.target.value)}
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
                </div>

                <div className="formSection mt-4">
                    <h3 className="sectionTitle">Your Profile</h3>
                    <p className="text-sm text-dim mb-4 mb-2">
                        You'll be added as the first member. Your email ({user?.email}) is linked automatically.
                    </p>
                    <FormField label="Your name" htmlFor="memName">
                        <input
                            id="memName"
                            className="input"
                            placeholder="e.g. Alex"
                            value={memberName}
                            onChange={(e) => setMemberName(e.target.value)}
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
                </div>

                <button
                    type="submit"
                    className="btn btnPrimary btnFull mt-6"
                    disabled={busy || !hhName.trim() || !memberName.trim()}
                >
                    {busy ? <><span className="spinIcon" aria-hidden>⟳</span> Creating…</> : 'Finish setup →'}
                </button>
            </form>
        </section>
    )
}
