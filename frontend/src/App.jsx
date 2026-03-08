import { useCallback, useEffect, useState } from 'react'
import { createExpense, deleteExpense, getHealth, getSettlement, listExpenses, listMembers, patchExpense } from './api'
import { AppHeader } from './components/AppHeader'
import { ExpenseForm } from './components/ExpenseForm'
import { ExpenseList } from './components/ExpenseList'
import { Banner } from './ui/Banner'
import { formatDollars } from './utils'

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
  const [members, setMembers] = useState([])

  // ── UI / loading flags ────────────────────────────────────────────────
  const [householdId, setHouseholdId]           = useState('home-main')
  const [loadingList, setLoadingList]           = useState(false)
  const [loadingMembers, setLoadingMembers]     = useState(false)
  const [loadingSettlement, setLoadingSettlement] = useState(false)
  const [submittingCreate, setSubmittingCreate] = useState(false)
  const [savingId, setSavingId]                 = useState('')
  const [deletingId, setDeletingId]             = useState('')

  // ── Feedback ──────────────────────────────────────────────────────────
  const [message, setMessage] = useState('')
  const [error, setError]     = useState('')
  const [settlement, setSettlement] = useState(null)
  const [settlementYear, setSettlementYear] = useState(new Date().getUTCFullYear())
  const [settlementMonth, setSettlementMonth] = useState(new Date().getUTCMonth() + 1)

  // ── Health check ──────────────────────────────────────────────────────
  useEffect(() => {
    let active = true
    getHealth()
      .then((status) => { if (active) setHealth(status === 'ok' ? 'ok' : status) })
      .catch(() => { if (active) setHealth('offline') })
    return () => { active = false }
  }, [])

  // ── Load expenses ─────────────────────────────────────────────────────
  const loadExpenses = useCallback(async () => {
    if (!householdId.trim()) {
      setError('household_id is required')
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
    if (!householdId.trim()) return
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
      setError(err.message)
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

  return (
    <div className="page">
      <AppHeader
        health={health}
        householdId={householdId}
        onHouseholdChange={setHouseholdId}
        onReload={loadExpenses}
        isLoading={isBusy}
      />

      {error   && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
      {message && <Banner type="ok"    onDismiss={() => setMessage('')}>{message}</Banner>}

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
        <div className="headerControls" style={{ marginTop: 12, marginBottom: 12 }}>
          <input
            className="householdInput"
            type="number"
            min="2000"
            max="2200"
            value={settlementYear}
            onChange={(e) => setSettlementYear(Number(e.target.value))}
            aria-label="Settlement year"
          />
          <input
            className="householdInput"
            type="number"
            min="1"
            max="12"
            value={settlementMonth}
            onChange={(e) => setSettlementMonth(Number(e.target.value))}
            aria-label="Settlement month"
          />
          <button type="button" className="btn btnGhost btnSm" onClick={loadSettlement} disabled={loadingSettlement}>
            {loadingSettlement ? 'Loading…' : 'Refresh settlement'}
          </button>
        </div>

        {settlement ? (
          <div className="formStack">
            <p>
              <strong>Total shared:</strong> {formatDollars(settlement.total_shared_cents)} | <strong>Mode:</strong>{' '}
              {settlement.effective_settlement_mode}
            </p>
            {settlement.fallback_reason ? <p>{settlement.fallback_reason}</p> : null}
            <p>
              Included expenses: {settlement.included_expense_count} | Excluded vouchers: {settlement.excluded_voucher_count}
            </p>
            <h3 className="sectionTitle">Transfers</h3>
            {Array.isArray(settlement.transfers) && settlement.transfers.length > 0 ? (
              <ul>
                {settlement.transfers.map((t, idx) => (
                  <li key={`${t.from_member_id}-${t.to_member_id}-${idx}`}>
                    {t.from_member_id} pays {t.to_member_id} {formatDollars(t.amount_cents)}
                  </li>
                ))}
              </ul>
            ) : (
              <p>No transfers needed.</p>
            )}
          </div>
        ) : (
          <p>Settlement unavailable for this period.</p>
        )}
      </section>
    </div>
  )
}

export default App
