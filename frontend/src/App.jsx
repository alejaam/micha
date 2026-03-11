import { useCallback, useEffect, useMemo, useState } from 'react'
import {
    createExpense,
    createHousehold,
    createMember,
    deleteExpense,
    getHealth,
    loginUser,
    patchExpense,
    registerUser,
    setAuthToken,
} from './api'
import { AppHeader } from './components/AppHeader'
import { AuthPanel } from './components/AuthPanel'
import { ExpenseForm } from './components/ExpenseForm'
import { ExpenseList } from './components/ExpenseList'
import { HouseholdSetupCard } from './components/HouseholdSetupCard'
import { MemberSetupCard } from './components/MemberSetupCard'
import { SettlementPanel } from './components/SettlementPanel'
import { useExpenses } from './hooks/useExpenses'
import { useHouseholds } from './hooks/useHouseholds'
import { useMembers } from './hooks/useMembers'
import { useSettlement } from './hooks/useSettlement'
import { Banner } from './ui/Banner'

const AUTH_STORAGE_KEY = 'micha_token'

function isExpectedSettlementOnboardingError(err) {
  return err?.code === 'NO_MEMBERS' || String(err?.message || '').toLowerCase().includes('at least one member')
}

function App() {
  const [token, setToken] = useState(() => localStorage.getItem(AUTH_STORAGE_KEY) ?? '')
  const [authMode, setAuthMode] = useState('login')
  const [authBusy, setAuthBusy] = useState(false)
  const [authError, setAuthError] = useState('')
  const [authMessage, setAuthMessage] = useState('')

  const isAuthenticated = token.trim() !== ''

  const [health, setHealth] = useState('checking...')
  const [submittingCreate, setSubmittingCreate] = useState(false)
  const [submittingHousehold, setSubmittingHousehold] = useState(false)
  const [submittingMember, setSubmittingMember] = useState(false)
  const [savingId, setSavingId] = useState('')
  const [deletingId, setDeletingId] = useState('')

  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const [newHouseholdName, setNewHouseholdName] = useState('')
  const [newSettlementMode, setNewSettlementMode] = useState('equal')
  const [newCurrency, setNewCurrency] = useState('MXN')
  const [newMemberName, setNewMemberName] = useState('')
  const [newMemberEmail, setNewMemberEmail] = useState('')
  const [newMemberSalary, setNewMemberSalary] = useState('0')

  const handleLogout = useCallback((reason = '') => {
    localStorage.removeItem(AUTH_STORAGE_KEY)
    setToken('')
    setAuthToken('')
    setAuthMode('login')
    setAuthBusy(false)
    setAuthMessage('')
    setAuthError(reason)
  }, [])

  const handleProtectedError = useCallback((err) => {
    if (err?.code === 'UNAUTHORIZED') {
      handleLogout('Session expired. Sign in again.')
      return true
    }

    setError(err.message)
    return false
  }, [handleLogout])

  const onErrorClear = useCallback(() => setError(''), [])
  const onUnexpectedError = useCallback((err) => setError(err.message), [])

  const {
    householdId,
    households,
    loadingHouseholds,
    setHouseholdId,
    setHouseholds,
    loadHouseholds,
  } = useHouseholds({
    isAuthenticated,
    handleProtectedError,
  })

  const {
    members,
    loadingMembers,
    setMembers,
    loadMembers,
  } = useMembers({
    isAuthenticated,
    householdId,
    handleProtectedError,
  })

  const {
    items,
    loadingList,
    setItems,
    loadExpenses,
  } = useExpenses({
    isAuthenticated,
    householdId,
    handleProtectedError,
    onErrorClear,
  })

  const {
    settlement,
    loadingSettlement,
    settlementYear,
    settlementMonth,
    setSettlement,
    setSettlementYear,
    setSettlementMonth,
    loadSettlement,
  } = useSettlement({
    isAuthenticated,
    householdId,
    handleProtectedError,
    onUnexpectedError,
    shouldIgnoreError: isExpectedSettlementOnboardingError,
  })

  const memberIndex = useMemo(
    () => Object.fromEntries(members.map((member) => [member.id, member.name])),
    [members],
  )

  const selectedHousehold = useMemo(
    () => households.find((household) => household.id === householdId) ?? null,
    [householdId, households],
  )

  const activeCurrency = selectedHousehold?.currency || 'MXN'

  const resetProtectedState = useCallback(() => {
    setHouseholdId('')
    setItems([])
    setHouseholds([])
    setMembers([])
    setSettlement(null)
    setMessage('')
    setError('')
    setSubmittingCreate(false)
    setSubmittingHousehold(false)
    setSubmittingMember(false)
    setSavingId('')
    setDeletingId('')
  }, [setHouseholdId, setItems, setHouseholds, setMembers, setSettlement])

  useEffect(() => {
    if (!isAuthenticated) {
      resetProtectedState()
    }
  }, [isAuthenticated, resetProtectedState])

  useEffect(() => {
    setAuthToken(token)
  }, [token])

  useEffect(() => {
    let active = true
    getHealth()
      .then((status) => {
        if (active) {
          setHealth(status === 'ok' ? 'ok' : status)
        }
      })
      .catch(() => {
        if (active) {
          setHealth('offline')
        }
      })

    return () => {
      active = false
    }
  }, [])

  async function handleLogin({ email, password }) {
    setAuthBusy(true)
    setAuthError('')
    setAuthMessage('')
    try {
      const out = await loginUser({ email, password })
      const nextToken = out?.token ?? ''
      if (!nextToken) {
        throw new Error('login succeeded but token was not returned')
      }

      localStorage.setItem(AUTH_STORAGE_KEY, nextToken)
      setAuthToken(nextToken)
      setToken(nextToken)
      setAuthError('')
    } catch (err) {
      setAuthError(err.message)
    } finally {
      setAuthBusy(false)
    }
  }

  async function handleRegister({ email, password }) {
    setAuthBusy(true)
    setAuthError('')
    setAuthMessage('')
    try {
      await registerUser({ email, password })
      setAuthMode('login')
      setAuthMessage('Account created. Sign in with your credentials.')
    } catch (err) {
      setAuthError(err.message)
    } finally {
      setAuthBusy(false)
    }
  }

  async function handleCreateHousehold(event) {
    event.preventDefault()
    if (!isAuthenticated || !newHouseholdName.trim()) {
      return
    }

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
      handleProtectedError(err)
    } finally {
      setSubmittingHousehold(false)
    }
  }

  async function handleCreateMember(event) {
    event.preventDefault()
    if (!isAuthenticated || !householdId || !newMemberName.trim() || !newMemberEmail.trim()) {
      return
    }

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
      handleProtectedError(err)
    } finally {
      setSubmittingMember(false)
    }
  }

  async function handleReloadAll() {
    if (!isAuthenticated) {
      return
    }

    await loadHouseholds()
    await loadMembers()
    await loadExpenses()
    await loadSettlement()
  }

  async function handleCreate({ amountCents, description, paidByMemberId, isShared, paymentMethod }) {
    if (!isAuthenticated) {
      return
    }

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
        currency: activeCurrency,
        paymentMethod,
      })
      setMessage('Expense added.')
      await loadExpenses()
      await loadSettlement()
    } catch (err) {
      handleProtectedError(err)
    } finally {
      setSubmittingCreate(false)
    }
  }

  async function handleSave({ id, amountCents, description }) {
    if (!isAuthenticated) {
      return
    }

    setMessage('')
    setError('')
    setSavingId(id)
    try {
      await patchExpense({ id, amountCents, description })
      setMessage('Expense updated.')
      await loadExpenses()
      await loadSettlement()
    } catch (err) {
      handleProtectedError(err)
    } finally {
      setSavingId('')
    }
  }

  async function handleDelete(id) {
    if (!isAuthenticated) {
      return
    }

    setMessage('')
    setError('')
    setDeletingId(id)
    try {
      await deleteExpense(id)
      setMessage('Expense deleted.')
      await loadExpenses()
      await loadSettlement()
    } catch (err) {
      handleProtectedError(err)
    } finally {
      setDeletingId('')
    }
  }

  const isBusy = submittingCreate || loadingList
  const hasHouseholds = households.length > 0
  const hasMembers = members.length > 0

  function handleAuthModeChange(nextMode) {
    setAuthMode(nextMode)
    setAuthError('')
    setAuthMessage('')
  }

  if (!isAuthenticated) {
    return (
      <main className="authShell">
        <AuthPanel
          mode={authMode}
          onModeChange={handleAuthModeChange}
          onLogin={handleLogin}
          onRegister={handleRegister}
          isSubmitting={authBusy}
          error={authError}
          message={authMessage}
        />
      </main>
    )
  }

  return (
    <div className="page">
      <AppHeader
        health={health}
        householdId={householdId}
        onHouseholdChange={setHouseholdId}
        onReload={handleReloadAll}
        onLogout={handleLogout}
        isLoading={isBusy}
        households={households}
      />

      {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
      {message && <Banner type="ok" onDismiss={() => setMessage('')}>{message}</Banner>}

      {!hasHouseholds ? (
        <HouseholdSetupCard
          newHouseholdName={newHouseholdName}
          onHouseholdNameChange={setNewHouseholdName}
          newSettlementMode={newSettlementMode}
          onSettlementModeChange={setNewSettlementMode}
          newCurrency={newCurrency}
          onCurrencyChange={setNewCurrency}
          onSubmit={handleCreateHousehold}
          isSubmitting={submittingHousehold}
          isLoading={loadingHouseholds}
        />
      ) : (
        !hasMembers ? (
          <MemberSetupCard
            newMemberName={newMemberName}
            onMemberNameChange={setNewMemberName}
            newMemberEmail={newMemberEmail}
            onMemberEmailChange={setNewMemberEmail}
            newMemberSalary={newMemberSalary}
            onMemberSalaryChange={setNewMemberSalary}
            onSubmit={handleCreateMember}
            isSubmitting={submittingMember}
            isLoading={loadingMembers}
          />
        ) : (
          <>
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
                currency={activeCurrency}
              />
            </div>

            <SettlementPanel
              settlement={settlement}
              settlementYear={settlementYear}
              settlementMonth={settlementMonth}
              onSettlementYearChange={setSettlementYear}
              onSettlementMonthChange={setSettlementMonth}
              onRefresh={loadSettlement}
              loadingSettlement={loadingSettlement}
              memberIndex={memberIndex}
              currency={activeCurrency}
            />
          </>
        )
      )}
    </div>
  )
}

export default App
