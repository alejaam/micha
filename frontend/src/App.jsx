import { useCallback, useEffect, useState } from 'react'
import { createExpense, deleteExpense, getHealth, listExpenses, patchExpense } from './api'
import { AppHeader } from './components/AppHeader'
import { ExpenseForm } from './components/ExpenseForm'
import { ExpenseList } from './components/ExpenseList'
import { Banner } from './ui/Banner'

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

  // ── UI / loading flags ────────────────────────────────────────────────
  const [householdId, setHouseholdId]           = useState('home-main')
  const [loadingList, setLoadingList]           = useState(false)
  const [submittingCreate, setSubmittingCreate] = useState(false)
  const [savingId, setSavingId]                 = useState('')
  const [deletingId, setDeletingId]             = useState('')

  // ── Feedback ──────────────────────────────────────────────────────────
  const [message, setMessage] = useState('')
  const [error, setError]     = useState('')

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

  // ── Create ────────────────────────────────────────────────────────────
  async function handleCreate({ amountCents, description }) {
    setMessage('')
    setError('')
    setSubmittingCreate(true)
    try {
      await createExpense({ householdId: householdId.trim(), amountCents, description })
      setMessage('Expense added.')
      await loadExpenses()
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
    </div>
  )
}

export default App
