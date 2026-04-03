import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getHousehold, listMembers, updateHousehold } from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { FormField } from '../ui/FormField'
import { isRecurringAutomationEnabled, setRecurringAutomationEnabled } from '../utils'

const CURRENCIES = [
    'MXN',
    'USD',
    'EUR',
    'COP',
    'ARS',
    'CLP',
    'PEN',
    'BRL',
]

export function HouseholdSettingsPage() {
    const navigate = useNavigate()
    const { handleProtectedError } = useAuth()
    const { householdId, selectedHousehold, loadHouseholds } = useAppShell()

    const [name, setName] = useState('')
    const [settlementMode, setSettlementMode] = useState('equal')
    const [currency, setCurrency] = useState('MXN')
    const [memberRows, setMemberRows] = useState([])
    const [recurringAutomationEnabled, setRecurringAutomationEnabledState] = useState(true)
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState('')
    const [message, setMessage] = useState('')

    useEffect(() => {
        async function loadSettings() {
            if (!householdId) return

            setLoading(true)
            setError('')
            try {
                const [household, members] = await Promise.all([
                    getHousehold({ householdId }),
                    listMembers({ householdId, limit: 100, offset: 0 }),
                ])

                setName(household?.name || selectedHousehold?.name || '')
                setSettlementMode(household?.settlement_mode || selectedHousehold?.settlement_mode || 'equal')
                setCurrency(household?.currency || selectedHousehold?.currency || 'MXN')

                const membersList = Array.isArray(members) ? members : []
                setMemberRows(membersList.map((member) => ({
                    id: member.id,
                    name: member.name,
                    monthlySalaryCents: Number(member.monthly_salary_cents) || 0,
                })))
                setRecurringAutomationEnabledState(isRecurringAutomationEnabled(householdId))
            } catch (err) {
                if (!handleProtectedError(err)) setError(err.message || 'Could not load household settings')
            } finally {
                setLoading(false)
            }
        }

        loadSettings()
    }, [handleProtectedError, householdId, selectedHousehold?.currency, selectedHousehold?.name, selectedHousehold?.settlement_mode])

    const settlementPreview = useMemo(() => {
        if (memberRows.length === 0) return []

        if (settlementMode === 'equal') {
            const base = Math.floor(1000 / memberRows.length) / 10
            const distributed = base * memberRows.length
            const remainder = Number((100 - distributed).toFixed(1))

            return memberRows.map((member, index) => ({
                memberId: member.id,
                memberName: member.name,
                percentage: (base + (index === 0 ? remainder : 0)).toFixed(1),
            }))
        }

        const totalSalary = memberRows.reduce((sum, member) => sum + member.monthlySalaryCents, 0)
        if (memberRows.length === 0 || totalSalary <= 0) return []

        return memberRows.map((member) => ({
            memberId: member.id,
            memberName: member.name,
            percentage: ((member.monthlySalaryCents / totalSalary) * 100).toFixed(1),
        }))
    }, [memberRows, settlementMode])

    const previewHint = settlementMode === 'proportional'
        ? 'Settlement uses salary-based proportional weights. Percentages are computed automatically.'
        : 'Settlement uses equal split. Percentages are distributed evenly and computed automatically.'

    const canSave = !!name.trim() && !saving && !loading

    function handleRecurringAutomationToggle(enabled) {
        setRecurringAutomationEnabledState(enabled)
        setRecurringAutomationEnabled(householdId, enabled)
    }

    async function handleSubmit(e) {
        e.preventDefault()
        if (!householdId || !canSave) return

        setSaving(true)
        setError('')
        setMessage('')
        try {
            await updateHousehold({
                householdId,
                name: name.trim(),
                settlementMode,
                currency,
            })

            await loadHouseholds()
            setMessage('Household settings updated.')
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message || 'Could not save household settings')
        } finally {
            setSaving(false)
        }
    }

    if (!householdId) {
        return (
            <section className="card" aria-label="Household settings unavailable">
                <Banner type="error">No household selected. Choose one first.</Banner>
                <button type="button" className="btn mt-4" onClick={() => navigate('/', { replace: true })}>
                    Back to dashboard
                </button>
            </section>
        )
    }

    return (
        <section className="card" aria-label="Household settings">
            <div className="listHeader mb-6">
                <div>
                    <h2 className="listTitle">Household settings</h2>
                    <p className="text-sm text-dim mt-1">
                        Update the setup values defined during onboarding.
                    </p>
                </div>
            </div>

            {error ? <Banner type="error">{error}</Banner> : null}
            {message ? <Banner type="ok">{message}</Banner> : null}

            <form className="formStack" onSubmit={handleSubmit}>
                <section className="formSection" aria-label="Member management">
                    <h3 className="sectionTitle">Member management</h3>
                    <p className="formHint">
                        Add or invite new members from this household settings area.
                    </p>
                    <button
                        type="button"
                        className="btn mt-4"
                        onClick={() => navigate('/members/new')}
                        disabled={loading || saving}
                    >
                        Add member
                    </button>
                </section>

                <section className="formSection" aria-label="Recurring automation policy">
                    <h3 className="sectionTitle">Recurring automation</h3>
                    <div className="sharedToggle">
                        <label htmlFor="settingsRecurringAutomation" className="sharedToggleLabel">
                            <input
                                id="settingsRecurringAutomation"
                                type="checkbox"
                                checked={recurringAutomationEnabled}
                                onChange={(e) => handleRecurringAutomationToggle(e.target.checked)}
                                disabled={loading || saving}
                            />
                            <span className="sharedToggleText">Auto-generate fixed expenses each month</span>
                        </label>
                        <p className="formHint">
                            When enabled, dashboard will generate recurring expenses once per period automatically.
                        </p>
                    </div>
                </section>

                <FormField label="Household name" htmlFor="settingsHouseholdName">
                    <input
                        id="settingsHouseholdName"
                        className="input"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        disabled={loading || saving}
                    />
                </FormField>

                <FormField label="Settlement mode" htmlFor="settingsSettlementMode">
                    <select
                        id="settingsSettlementMode"
                        className="input"
                        value={settlementMode}
                        onChange={(e) => setSettlementMode(e.target.value)}
                        disabled={loading || saving}
                    >
                        <option value="equal">Equal split</option>
                        <option value="proportional">Proportional to salary</option>
                    </select>
                </FormField>

                <FormField label="Currency" htmlFor="settingsCurrency">
                    <select
                        id="settingsCurrency"
                        className="input"
                        value={currency}
                        onChange={(e) => setCurrency(e.target.value)}
                        disabled={loading || saving}
                    >
                        {CURRENCIES.map((code) => (
                            <option key={code} value={code}>{code}</option>
                        ))}
                    </select>
                </FormField>

                <section className="formSection mt-4" aria-label="Settlement distribution preview">
                    <h3 className="sectionTitle">Settlement distribution preview</h3>
                    <p className="formHint">{previewHint}</p>
                    {settlementPreview.length > 0 ? (
                        <div className="formStack">
                            {settlementPreview.map((row) => (
                                <p key={row.memberId} className="formHint">
                                    <strong>{row.memberName}</strong>: {row.percentage}%
                                </p>
                            ))}
                        </div>
                    ) : (
                        <p className="formHint formHintWarning">
                            Preview unavailable. Add members and salary data to visualize distribution.
                        </p>
                    )}
                </section>

                <div className="flex gap-4 mt-6">
                    <button
                        type="button"
                        className="btn flex-1"
                        onClick={() => navigate('/', { replace: true })}
                        disabled={saving}
                    >
                        Back
                    </button>
                    <button
                        type="submit"
                        className="btn btnPrimary flex-1"
                        disabled={!canSave}
                    >
                        {saving ? 'Saving...' : 'Save settings'}
                    </button>
                </div>
            </form>
        </section>
    )
}
