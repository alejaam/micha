import { useCallback, useEffect, useMemo, useState } from 'react'
import { createExpense, createHousehold, createMember, deleteExpense, getHealth, getSettlement, listExpenses, listHouseholds, listMembers, patchExpense } from './api'
import { AppHeader } from './components/AppHeader'
import { ExpenseForm } from './components/ExpenseForm'
import { ExpenseList } from './components/ExpenseList'
import { Banner } from './ui/Banner'
import { FormField } from './ui/FormField'
import { formatDollars } from './utils'

const SETTLEMENT_YEARS = Array.from({ length: 6 }, (_, i) => new Date().getUTCFullYear() - 4 + i)
const SETTLEMENT_MONTHS = [
  { value: 1,  label: '01 · Jan' }, { value: 2,  label: '02 · Feb' },
  { value: 3,  label: '03 · Mar' }, { value: 4,  label: '04 · Apr' },
  { value: 5,  label: '05 · May' }, { value: 6,  label: '06 · Jun' },
  { value: 7,  label: '07 · Jul' }, { value: 8,  label: '08 · Aug' },
  { value: 9,  label: '09 · Sep' }, { value: 10, label: '10 · Oct' },
  { value: 11, label: '11 · Nov' }, { value: 12, label: '12 · Dec' },
]

function isExpectedSettlementOnboardingError(err) {
  return err?.code === 'NO_MEMBERS' || String(err?.message || '').toLowerCase().includes('at least one member')
}

/**
 * App — root state orchestrator.
 *
 * All server state lives here and is passed down as props.
 * UI sub-trees are handled by their respective components.
 */
function App() {
  // ── Server state ────────────────────────────────────────────────────────
  const [health, setHealth] = useState('checking…')
  const [items, setItems]   = useState([])
  const [households, setHouseholds] = useState([])
  const [members, setMembers] = useState([])

  // ── UI / loading flags ────────────────────────────────────────────────
  const [householdId, setHouseholdId]           = useState('')
  const [loadingHouseholds, setLoadingHouseholds] = useState(false)
  const [loadingList, setLoadingList]           = useState(false)
  const [loadingMembers, setLoadingMembers]     = useState(false)
  const [loadingSettlement, setLoadingSettlement] = useState(false)
  const [submittingCreate, setSubmittingCreate] = useState(false)
  const [submittingHousehold, setSubmittingHousehold] = useState(false)
  const [submittingMember, setSubmittingMember] = useState(false)
  const [savingId, setSavingId]                 = useState('')
  const [deletingId, setDeletingId]             = useState('')

  // ── Feedback ──────────────────────────────────────────────────────────
  const [message, setMessage] = useState('')
  const [error, setError]     = useState('')
  const [settlement, setSettlement] = useState(null)
  const [settlementYear, setSettlementYear] = useState(new Date().getUTCFullYear())
  const [settlementMonth, setSettlementMonth] = useState(new Date().getUTCMonth() + 1)
  const [newHouseholdName, setNewHouseholdName] = useState('')
  const [newSettlementMode, setNewSettlementMode] = useState('equal')
  const [newCurrency, setNewCurrency] = useState('MXN')
  const [newMemberName, setNewMemberName] = useState('')
  const [newMemberEmail, setNewMemberEmail] = useState('')
  const [newMemberSalary, setNewMemberSalary] = useState('0')
  // ── Derived ──────────────────────────────────────────────────────────
  const memberIndex = useMemo(
    () => Object.fromEntries(members.map((m) => [m.id, m.name])),
    [members],
  )
  // ── Health check ──────────────────────────────────────────────────────
  useEffect(() => {
    let active = true
    getHealth()
      .then((status) => { if (active) setHealth(status === 'ok' ? 'ok' : status) })
      .catch(() => { if (active) setHealth('offline') })
    return () => { active = false }
  }, [])

  const loadHouseholds = useCallback(async () => {
    setLoadingHouseholds(true)
    try {
      const data = await listHouseholds({ limit: 100, offset: 0 })
      const next = Array.isArray(data) ? data : []
      setHouseholds(next)

      if (next.length === 0) {
        setHouseholdId('')
        setItems([])
        setMembers([])
        setSettlement(null)
      } else {
        const selectedExists = next.some((h) => h.id === householdId)
        if (!selectedExists) {
          setHouseholdId(next[0].id)
        }
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoadingHouseholds(false)
    }
  }, [householdId])

  useEffect(() => { loadHouseholds() }, [loadHouseholds])

  // ── Load expenses ─────────────────────────────────────────────────────
  const loadExpenses = useCallback(async () => {
    if (!householdId.trim()) {
      setItems([])
      return
    }
    setLoadingList(true)
    setError('')
    try {
      const data = await listExpenses({ householdId: householdId.trim(), limit: 50, offset: 0 })
      setItems(Array.isArray(data) ? data : [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoadingList(false)
    }
  }, [householdId])

  useEffect(() => { loadExpenses() }, [loadExpenses])

  const loadSettlement = useCallback(async () => {
    if (!householdId.trim()) {
      setSettlement(null)
      return
    }
    setLoadingSettlement(true)
    try {
      const data = await getSettlement({
        householdId: householdId.trim(),
        year: settlementYear,
        month: settlementMonth,
      })
      setSettlement(data)
    } catch (err) {
      setSettlement(null)
      if (!isExpectedSettlementOnboardingError(err)) {
        setError(err.message)
      }
    } finally {
      setLoadingSettlement(false)
    }
  }, [householdId, settlementMonth, settlementYear])

  useEffect(() => { loadSettlement() }, [loadSettlement])

  const loadMembers = useCallback(async () => {
    if (!householdId.trim()) {
      setMembers([])
      return
    }

    setLoadingMembers(true)
    try {
      const data = await listMembers({ householdId: householdId.trim(), limit: 100, offset: 0 })
      setMembers(Array.isArray(data) ? data : [])
    } catch (err) {
      setMembers([])
      setError(err.message)
    } finally {
      setLoadingMembers(false)
    }
  }, [householdId])

  useEffect(() => { loadMembers() }, [loadMembers])

  async function handleCreateHousehold(event) {
    event.preventDefault()
    if (!newHouseholdName.trim()) return

    setSubmittingHousehold(true)
    setError('')
    setMessage('')
    try {
      const out = await createHousehold({
        name: newHouseholdName.trim(),
        settlementMode: newSettlementMode,
        currency: newCurrency.trim().toUpperCase() || 'MXN',
      })

      setMessage('Household created.')
      setNewHouseholdName('')
      await loadHouseholds()
      if (out?.household_id) {
        setHouseholdId(out.household_id)
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmittingHousehold(false)
    }
  }

  async function handleCreateMember(event) {
    event.preventDefault()
    if (!householdId || !newMemberName.trim() || !newMemberEmail.trim()) return

    setSubmittingMember(true)
    setError('')
    setMessage('')
    try {
      await createMember({
        householdId,
        name: newMemberName.trim(),
        email: newMemberEmail.trim(),
        monthlySalaryCents: Number(newMemberSalary) || 0,
      })
      setMessage('Member created.')
      setNewMemberName('')
      setNewMemberEmail('')
      setNewMemberSalary('0')
      await loadMembers()
      await loadSettlement()
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmittingMember(false)
    }
  }

  async function handleReloadAll() {
    await loadHouseholds()
    await loadMembers()
    await loadExpenses()
    await loadSettlement()
  }

  // ── Create ────────────────────────────────────────────────────────────
  async function handleCreate({ amountCents, description, paidByMemberId, isShared, paymentMethod }) {
    setMessage('')
    setError('')
    setSubmittingCreate(true)
    try {
      await createExpense({
        householdId: householdId.trim(),
        paidByMemberId,
        amountCents,
        description,
        isShared,
        paymentMethod,
      })
      setMessage('Expense added.')
      await loadExpenses()
      await loadSettlement()
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmittingCreate(false)
    }
  }

  // ── Save (patch) ──────────────────────────────────────────────────────
  async function handleSave({ id, amountCents, description }) {
    setMessage('')
    setError('')
    setSavingId(id)
    try {
      await patchExpense({ id, amountCents, description })
      setMessage('Expense updated.')
      await loadExpenses()
      await loadSettlement()
    } catch (err) {
      setError(err.message)
    } finally {
      setSavingId('')
    }
  }

  // ── Delete ────────────────────────────────────────────────────────────
  async function handleDelete(id) {
    setMessage('')
    setError('')
    setDeletingId(id)
    try {
      await deleteExpense(id)
      setMessage('Expense deleted.')
      await loadExpenses()
      await loadSettlement()
    } catch (err) {
      setError(err.message)
    } finally {
      setDeletingId('')
    }
  }

  // ── Render ────────────────────────────────────────────────────────────
  const isBusy = submittingCreate || loadingList
  const hasHouseholds = households.length > 0
  const hasMembers = members.length > 0

  return (
    <div className="page">
      <AppHeader
        health={health}
        householdId={householdId}
        onHouseholdChange={setHouseholdId}
        onReload={handleReloadAll}
        isLoading={isBusy}
        households={households}
      />

      {error   && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
      {message && <Banner type="ok"    onDismiss={() => setMessage('')}>{message}</Banner>}

      {!hasHouseholds ? (
        <section className="card" aria-label="Create first household">
          <h2 className="sectionTitle">Create your first household</h2>
          <form className="formStack" onSubmit={handleCreateHousehold}>
            <FormField label="Name" htmlFor="newHouseholdName">
              <input
                id="newHouseholdName"
                className="input"
                placeholder="e.g. Casa Familia"
                value={newHouseholdName}
                onChange={(e) => setNewHouseholdName(e.target.value)}
                disabled={submittingHousehold || loadingHouseholds}
              />
            </FormField>
            <FormField label="Settlement mode" htmlFor="newSettlementMode">
              <select
                id="newSettlementMode"
                className="input"
                value={newSettlementMode}
                onChange={(e) => setNewSettlementMode(e.target.value)}
                disabled={submittingHousehold || loadingHouseholds}
              >
                <option value="equal">Equal split</option>
                <option value="proportional">Proportional to salary</option>
              </select>
            </FormField>
            <FormField label="Currency" htmlFor="newCurrency">
              <input
                id="newCurrency"
                className="input"
                placeholder="MXN"
                value={newCurrency}
                onChange={(e) => setNewCurrency(e.target.value)}
                disabled={submittingHousehold || loadingHouseholds}
              />
            </FormField>
            <button type="submit" className="btn btnPrimary" disabled={submittingHousehold || loadingHouseholds || !newHouseholdName.trim()}>
              {submittingHousehold ? <><span className="spinIcon" aria-hidden>&#x27F3;</span> Creating…</> : 'Create household'}
            </button>
          </form>
        </section>
      ) : (
        <>
          {!hasMembers ? (
            <section className="card" aria-label="Create first member">
              <h2 className="sectionTitle">Create your first member</h2>
              <form className="formStack" onSubmit={handleCreateMember}>
                <FormField label="Name" htmlFor="newMemberName">
                  <input
                    id="newMemberName"
                    className="input"
                    placeholder="e.g. Alex"
                    value={newMemberName}
                    onChange={(e) => setNewMemberName(e.target.value)}
                    disabled={submittingMember || loadingMembers}
                  />
                </FormField>
                <FormField label="Email" htmlFor="newMemberEmail">
                  <input
                    id="newMemberEmail"
                    className="input"
                    type="email"
                    placeholder="alex@example.com"
                    value={newMemberEmail}
                    onChange={(e) => setNewMemberEmail(e.target.value)}
                    disabled={submittingMember || loadingMembers}
                  />
                </FormField>
                <FormField label="Monthly salary (cents)" htmlFor="newMemberSalary">
                  <input
                    id="newMemberSalary"
                    className="input"
                    type="number"
                    min="0"
                    placeholder="0"
                    value={newMemberSalary}
                    onChange={(e) => setNewMemberSalary(e.target.value)}
                    disabled={submittingMember || loadingMembers}
                  />
                </FormField>
                <button
                  type="submit"
                  className="btn btnPrimary"
                  disabled={submittingMember || loadingMembers || !newMemberName.trim() || !newMemberEmail.trim()}
                >
                  {submittingMember ? <><span className="spinIcon" aria-hidden>&#x27F3;</span> Creating…</> : 'Create member'}
                </button>
              </form>
            </section>
          ) : null}

          <div className="pageGrid">
            <ExpenseForm
              onSubmit={handleCreate}
              isSubmitting={submittingCreate}
              members={members}
              isLoadingMembers={loadingMembers}
            />
            <ExpenseList
              items={items}
              isLoading={loadingList}
              deletingId={deletingId}
              savingId={savingId}
              onDelete={handleDelete}
              onSave={handleSave}
            />
          </div>

          <section className="card" aria-label="Monthly settlement">
            <h2 className="sectionTitle">Monthly settlement</h2>
            <div className="settlementControls">
              <div className="householdRow">
                <label htmlFor="settlementYear" className="householdLabel">Year</label>
                <select
                  id="settlementYear"
                  className="input inputSm"
                  style={{ width: 90 }}
                  value={settlementYear}
                  onChange={(e) => setSettlementYear(Number(e.target.value))}
                  aria-label="Settlement year"
                >
                  {SETTLEMENT_YEARS.map((y) => (
                    <option key={y} value={y}>{y}</option>
                  ))}
                </select>
              </div>
              <div className="householdRow">
                <label htmlFor="settlementMonth" className="householdLabel">Month</label>
                <select
                  id="settlementMonth"
                  className="input inputSm"
                  style={{ width: 120 }}
                  value={settlementMonth}
                  onChange={(e) => setSettlementMonth(Number(e.target.value))}
                  aria-label="Settlement month"
                >
                  {SETTLEMENT_MONTHS.map(({ value, label }) => (
                    <option key={value} value={value}>{label}</option>
                  ))}
                </select>
              </div>
              <button type="button" className="btn btnGhost btnSm" onClick={loadSettlement} disabled={loadingSettlement}>
                {loadingSettlement ? <><span className="spinIcon" aria-hidden>&#x27F3;</span> Loading…</> : 'Refresh'}
              </button>
            </div>

            {settlement ? (
              <div className="formStack">
                {settlement.fallback_reason ? (
                  <p className="settlementFallback" role="alert">⚠ {settlement.fallback_reason}</p>
                ) : null}
                <div className="settlementStats">
                  <span className="settlementStat">
                    <span className="settlementStatLabel">Total shared:</span>
                    <span className="settlementStatValue">{formatDollars(settlement.total_shared_cents)}</span>
                  </span>
                  <span className="settlementStat">
                    <span className="settlementStatLabel">Mode:</span>
                    <span className="settlementStatValue">{settlement.effective_settlement_mode}</span>
                  </span>
                  <span className="settlementStat">
                    <span className="settlementStatLabel">Expenses:</span>
                    <span className="settlementStatValue">{settlement.included_expense_count}</span>
                  </span>
                  <span className="settlementStat">
                    <span className="settlementStatLabel">Excluded:</span>
                    <span className="settlementStatValue">{settlement.excluded_voucher_count}</span>
                  </span>
                </div>
                <h3 className="sectionTitle">
                  <span className="sectionTitleIcon" aria-hidden>⇄</span>
                  Transfers
                </h3>
                {Array.isArray(settlement.transfers) && settlement.transfers.length > 0 ? (
                  <ul className="transferList">
                    {settlement.transfers.map((t, idx) => (
                      <li key={`${t.from_member_id}-${t.to_member_id}-${idx}`} className="transferItem">
                        <span className="transferNames">
                          {memberIndex[t.from_member_id] ?? t.from_member_id.slice(0, 8) + '…'}
                          <span className="transferArrow" aria-hidden>→</span>
                          {memberIndex[t.to_member_id] ?? t.to_member_id.slice(0, 8) + '…'}
                        </span>
                        <span className="transferAmount">{formatDollars(t.amount_cents)}</span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="emptyHint">No transfers needed — everyone is settled!</p>
                )}
              </div>
            ) : (
              <div className="emptyState">
                <div className="emptyIcon" aria-hidden>📊</div>
                <p className="emptyTitle">No settlement data</p>
                <p className="emptyHint">No expenses recorded for this period.</p>
              </div>
            )}
          </section>
        </>
      )}
    </div>
  )
}

export default App
