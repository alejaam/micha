import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createHousehold, createMember } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'
import { dollarsToCents } from '../utils'

const CURRENCIES = [
    { code: 'MXN', label: '🇲🇽 MXN — Mexican Peso' },
    { code: 'USD', label: '🇺🇸 USD — US Dollar' },
    { code: 'EUR', label: '🇪🇺 EUR — Euro' },
    { code: 'COP', label: '🇨🇴 COP — Colombian Peso' },
    { code: 'ARS', label: '🇦🇷 ARS — Argentine Peso' },
    { code: 'CLP', label: '🇨🇱 CLP — Chilean Peso' },
    { code: 'PEN', label: '🇵🇪 PEN — Peruvian Sol' },
    { code: 'BRL', label: '🇧🇷 BRL — Brazilian Real' },
]

const SETTLEMENT_HINTS = {
    equal: 'Each member pays the same share regardless of income.',
    proportional: 'Members who earn more contribute a larger share of expenses.',
}

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
    const [salaryDollars, setSalaryDollars] = useState('')

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
                currency,
            })
            
            const createdHouseholdId = hhOut?.household_id ?? hhOut?.id ?? ''
            if (!createdHouseholdId) {
                throw new Error('household created but id was not returned')
            }

            // Keep the new ID locally
            setHouseholdId(createdHouseholdId)

            // 2. Auto-create the creator as the first member
            const salaryCents = dollarsToCents(salaryDollars) || 0
            await createMember({
                householdId: createdHouseholdId,
                name: memberName.trim(),
                email: user?.email || '',
                monthlySalaryCents: salaryCents,
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
                            autoFocus
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
                        <p className="formHint">{SETTLEMENT_HINTS[settlementMode]}</p>
                    </FormField>
                    <FormField label="Currency" htmlFor="hhCurrency">
                        <select
                            id="hhCurrency"
                            className="input"
                            value={currency}
                            onChange={(e) => setCurrency(e.target.value)}
                            disabled={busy}
                        >
                            {CURRENCIES.map((c) => (
                                <option key={c.code} value={c.code}>{c.label}</option>
                            ))}
                        </select>
                    </FormField>
                </div>

                <div className="formSection mt-4">
                    <h3 className="sectionTitle">Your Profile</h3>
                    <p className="text-sm text-dim mb-2">
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
                    <FormField label="Monthly salary (optional)" htmlFor="memSalary">
                        <div className="inputWrap">
                            <span className="inputPrefix" aria-hidden>$</span>
                            <input
                                id="memSalary"
                                className="input inputWithPrefix"
                                type="number"
                                min="0"
                                step="0.01"
                                placeholder="e.g. 30000"
                                value={salaryDollars}
                                onChange={(e) => setSalaryDollars(e.target.value)}
                                disabled={busy}
                            />
                        </div>
                        <p className="formHint">Used to calculate proportional splits. You can update this later.</p>
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
