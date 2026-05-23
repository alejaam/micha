import { motion } from 'framer-motion'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { SubscriptionCatalogSelector } from '../components/SubscriptionCatalogSelector'
import { SubscriptionKPIPanel } from '../components/SubscriptionKPIPanel'
import { useAppShell } from '../context/AppShellContext'
import { useRecurringExpensesCRUD } from '../hooks/useRecurringExpensesCRUD'
import { Banner } from '../ui/Banner'
import { EmptyState } from '../ui/EmptyState'
import { listExpenseCatalogLinks } from '../api'
import { formatCurrency } from '../utils'

const EXPENSE_TYPE_LABELS = {
    fixed: 'Fijo',
    variable: 'Variable',
    msi: 'MSI',
}

const RECURRENCE_LABELS = {
    monthly: 'Mensual',
    biweekly: 'Quincenal',
    weekly: 'Semanal',
}

function defaultFormPayload(householdId) {
    const today = new Date().toISOString().slice(0, 10)
    return {
        householdId,
        paidByMemberId: '',
        isAgnostic: true,
        amountCents: 0,
        description: '',
        category: 'other',
        expenseType: 'fixed',
        recurrencePattern: 'monthly',
        startDate: today,
        endDate: null,
    }
}

/**
 * FixedExpensesPage — standalone page for managing recurring fixed expenses.
 * Desktop: sortable table. Mobile (<640px): stacked cards.
 */
export function FixedExpensesPage() {
    const { householdId } = useAppShell()
    const {
        items,
        loading,
        error,
        loadItems,
        createItem,
        updateItem,
        deleteItem,
        editingId,
        setEditingId,
    } = useRecurringExpensesCRUD({ householdId })

    const [showForm, setShowForm] = useState(false)
    const [formData, setFormData] = useState(() => defaultFormPayload(householdId))
    const [submitting, setSubmitting] = useState(false)
    const [localError, setLocalError] = useState(null)
    const [catalogExpenseId, setCatalogExpenseId] = useState(null)
    const [catalogLinkedIds, setCatalogLinkedIds] = useState([])
    const [showKPIPanel, setShowKPIPanel] = useState(false)

    // Reset form when household changes
    useEffect(() => {
        setFormData(defaultFormPayload(householdId))
    }, [householdId])

    const handleFormChange = useCallback((field, value) => {
        setFormData((prev) => ({ ...prev, [field]: value }))
    }, [])

    const handleSubmit = useCallback(async (event) => {
        event?.preventDefault()
        if (formData.amountCents <= 0) {
            setLocalError('El monto debe ser mayor a cero')
            return
        }
        if (!formData.description.trim()) {
            setLocalError('La descripción es obligatoria')
            return
        }
        setSubmitting(true)
        setLocalError(null)
        try {
            const result = await createItem({
                householdId: formData.householdId,
                paidByMemberId: formData.paidByMemberId,
                isAgnostic: formData.isAgnostic,
                amountCents: formData.amountCents,
                description: formData.description.trim(),
                category: formData.category,
                expenseType: formData.expenseType,
                recurrencePattern: formData.recurrencePattern,
                startDate: formData.startDate,
                endDate: formData.endDate,
            })
            if (result) {
                setShowForm(false)
                setFormData(defaultFormPayload(householdId))
            }
        } finally {
            setSubmitting(false)
        }
    }, [formData, createItem, householdId])

    const handleEdit = useCallback((item) => {
        setEditingId(item.id)
        setFormData({
            householdId: item.household_id,
            paidByMemberId: item.paid_by_member_id ?? '',
            isAgnostic: item.is_agnostic ?? false,
            amountCents: item.amount_cents,
            description: item.description,
            category: item.category_id ?? 'other',
            expenseType: item.expense_type ?? 'fixed',
            recurrencePattern: item.recurrence_pattern ?? 'monthly',
            startDate: item.start_date,
            endDate: item.end_date ?? null,
        })
        setShowForm(true)
    }, [setEditingId])

    const handleUpdate = useCallback(async () => {
        if (formData.amountCents <= 0) {
            setLocalError('El monto debe ser mayor a cero')
            return
        }
        if (!formData.description.trim()) {
            setLocalError('La descripción es obligatoria')
            return
        }
        setSubmitting(true)
        setLocalError(null)
        try {
            const success = await updateItem(editingId, {
                amountCents: formData.amountCents,
                description: formData.description.trim(),
                category: formData.category,
                recurrencePattern: formData.recurrencePattern,
                startDate: formData.startDate,
                endDate: formData.endDate,
            })
            if (success) {
                setShowForm(false)
                setFormData(defaultFormPayload(householdId))
            }
        } finally {
            setSubmitting(false)
        }
    }, [editingId, formData, updateItem, householdId])

    const handleDelete = useCallback(async (item) => {
        if (!window.confirm(`¿Eliminar "${item.description}"?`)) return
        await deleteItem(item.id)
    }, [deleteItem])

    const handleOpenCatalog = useCallback(async (item) => {
        setCatalogExpenseId(item.id)
        try {
            const links = await listExpenseCatalogLinks({ recurringExpenseId: item.id })
            setCatalogLinkedIds(Array.isArray(links) ? links.map((l) => l.catalog_service_id) : [])
        } catch {
            setCatalogLinkedIds([])
        }
    }, [])

    const handleCloseCatalog = useCallback(() => {
        setCatalogExpenseId(null)
        setCatalogLinkedIds([])
    }, [])

    const handleCatalogSave = useCallback(() => {
        // Refresh the expense list to get updated state
        loadItems()
    }, [loadItems])

    const handleCancel = useCallback(() => {
        setShowForm(false)
        setEditingId(null)
        setFormData(defaultFormPayload(householdId))
        setLocalError(null)
    }, [setEditingId, householdId])

    const displayError = error || localError

    // Sortable columns logic
    const [sortField, setSortField] = useState('created_at')
    const [sortDir, setSortDir] = useState('desc')

    const sortedItems = useMemo(() => {
        const sorted = [...items]
        sorted.sort((a, b) => {
            let cmp = 0
            switch (sortField) {
                case 'description':
                    cmp = (a.description ?? '').localeCompare(b.description ?? '')
                    break
                case 'amount_cents':
                    cmp = (a.amount_cents ?? 0) - (b.amount_cents ?? 0)
                    break
                case 'expense_type':
                    cmp = (a.expense_type ?? '').localeCompare(b.expense_type ?? '')
                    break
                case 'recurrence_pattern':
                    cmp = (a.recurrence_pattern ?? '').localeCompare(b.recurrence_pattern ?? '')
                    break
                default:
                    cmp = new Date(a.created_at ?? 0) - new Date(b.created_at ?? 0)
            }
            return sortDir === 'asc' ? cmp : -cmp
        })
        return sorted
    }, [items, sortField, sortDir])

    const handleSort = useCallback((field) => {
        setSortField((prev) => {
            if (prev === field) {
                setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'))
                return prev
            }
            setSortDir('asc')
            return field
        })
    }, [])

    const SortIcon = ({ field }) => {
        if (sortField !== field) return <span className="sortIcon"> ↕</span>
        return <span className="sortIcon">{sortDir === 'asc' ? ' ↑' : ' ↓'}</span>
    }

    return (
        <motion.div
            className="pageGrid"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.2 }}
        >
            {displayError && (
                <Banner type="error" floating onDismiss={() => { setLocalError(null); /* parent error remains */ }}>
                    {displayError}
                </Banner>
            )}

            <section className="card spanAll" aria-label="Gastos fijos">
                <div className="listHeader">
                    <h2 className="listTitle">Gastos fijos</h2>
                    {items.length > 0 && (
                        <span className="listCount">{items.length} registrados</span>
                    )}
                </div>

                {!showForm && (
                    <div className="u-flex u-mb-3">
                        <button
                            type="button"
                            className="btn btnPrimary"
                            onClick={() => {
                                setFormData(defaultFormPayload(householdId))
                                setEditingId(null)
                                setShowForm(true)
                            }}
                        >
                            + Nuevo gasto fijo
                        </button>
                    </div>
                )}

                {/* ─── Create / Edit form ─── */}
                {showForm && (
                    <form onSubmit={editingId ? (e) => { e.preventDefault(); handleUpdate() } : handleSubmit} className="formStack">
                        <div className="formField">
                            <label className="formLabel" htmlFor="fe-description">Descripción</label>
                            <input
                                id="fe-description"
                                className="input"
                                type="text"
                                value={formData.description}
                                onChange={(e) => handleFormChange('description', e.target.value)}
                                required
                                placeholder="Ej: Renta, Netflix..."
                            />
                        </div>
                        <div className="formField">
                            <label className="formLabel" htmlFor="fe-amount">Monto (centavos)</label>
                            <input
                                id="fe-amount"
                                className="input"
                                type="number"
                                min="1"
                                step="1"
                                value={formData.amountCents}
                                onChange={(e) => handleFormChange('amountCents', Number(e.target.value) || 0)}
                                required
                            />
                        </div>
                        <div className="formField">
                            <label className="formLabel" htmlFor="fe-pattern">Frecuencia</label>
                            <select
                                id="fe-pattern"
                                className="input"
                                value={formData.recurrencePattern}
                                onChange={(e) => handleFormChange('recurrencePattern', e.target.value)}
                            >
                                <option value="monthly">Mensual</option>
                                <option value="biweekly">Quincenal</option>
                                <option value="weekly">Semanal</option>
                            </select>
                        </div>
                        <div className="formField">
                            <label className="formLabel" htmlFor="fe-start">Fecha inicio</label>
                            <input
                                id="fe-start"
                                className="input"
                                type="date"
                                value={formData.startDate}
                                onChange={(e) => handleFormChange('startDate', e.target.value)}
                                required
                            />
                        </div>
                        <div className="formField">
                            <label className="formLabel" htmlFor="fe-agnostic">
                                <input
                                    id="fe-agnostic"
                                    type="checkbox"
                                    checked={formData.isAgnostic}
                                    onChange={(e) => handleFormChange('isAgnostic', e.target.checked)}
                                    style={{ marginRight: '0.5rem' }}
                                />
                                Sin pagador específico
                            </label>
                            <p className="u-text-xs u-text-dim">Se divide entre todos los miembros</p>
                        </div>
                        <div className="u-flex u-gap-2 u-mt-2">
                            <button type="submit" className="btn btnPrimary" disabled={submitting}>
                                {submitting ? 'Guardando...' : editingId ? 'Guardar cambios' : 'Crear gasto fijo'}
                            </button>
                            <button type="button" className="btn btnGhost" onClick={handleCancel}>
                                Cancelar
                            </button>
                        </div>
                    </form>
                )}

                {/* ─── Loading state ─── */}
                {loading && (
                    <div className="emptyState">
                        <p className="emptyTitle">Cargando...</p>
                    </div>
                )}

                {/* ─── Empty state ─── */}
                {!loading && sortedItems.length === 0 && !showForm && (
                    <EmptyState
                        title="Sin gastos fijos aún"
                        description="Crea tu primer gasto fijo para empezar a dar seguimiento."
                        icon="#"
                    />
                )}

                {/* ─── Desktop table (≥640px) ─── */}
                {sortedItems.length > 0 && (
                    <div className="fixedExpensesTableWrapper">
                        <table className="fixedExpensesTable">
                            <thead>
                                <tr>
                                    <th onClick={() => handleSort('description')} className="sortable col-description">
                                        Descripción<SortIcon field="description" />
                                    </th>
                                    <th onClick={() => handleSort('amount_cents')} className="sortable col-amount">
                                        Monto<SortIcon field="amount_cents" />
                                    </th>
                                    <th onClick={() => handleSort('expense_type')} className="sortable col-type">
                                        Tipo<SortIcon field="expense_type" />
                                    </th>
                                    <th onClick={() => handleSort('recurrence_pattern')} className="sortable col-frequency">
                                        Frecuencia<SortIcon field="recurrence_pattern" />
                                    </th>
                                    <th className="col-status">Estado</th>
                                    <th className="col-actions">Acciones</th>
                                </tr>
                            </thead>
                            <tbody>
                                {sortedItems.map((item) => (
                                    <tr key={item.id}>
                                    <td className="col-description">{item.description}</td>
                                    <td className="col-amount">{formatCurrency(item.amount_cents, 'MXN')}</td>
                                    <td className="col-type">{EXPENSE_TYPE_LABELS[item.expense_type] ?? item.expense_type}</td>
                                    <td className="col-frequency">{RECURRENCE_LABELS[item.recurrence_pattern] ?? item.recurrence_pattern}</td>
                                    <td className="col-status">
                                        <span className={`statusBadge ${item.is_active ? 'statusActive' : 'statusInactive'}`}>
                                            {item.is_active ? 'Activo' : 'Inactivo'}
                                        </span>
                                    </td>
                                    <td className="col-actions">
                                        <button
                                            type="button"
                                            className="btn btnGhost btnSm btnIcon"
                                            title="Asociar servicios"
                                            onClick={() => handleOpenCatalog(item)}
                                            aria-label={`Asociar servicios a ${item.description}`}
                                        >
                                            🔗
                                        </button>
                                        <button
                                            type="button"
                                            className="btn btnGhost btnSm btnIcon"
                                            title="Editar"
                                            onClick={() => handleEdit(item)}
                                            aria-label={`Editar ${item.description}`}
                                        >
                                            ✏️
                                        </button>
                                        <button
                                            type="button"
                                            className="btn btnGhost btnSm btnIcon"
                                            title="Eliminar"
                                            onClick={() => handleDelete(item)}
                                            aria-label={`Eliminar ${item.description}`}
                                        >
                                            🗑️
                                        </button>
                                    </td>
                                </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}

                {/* ─── Mobile cards (<640px) ─── */}
                {sortedItems.length > 0 && (
                    <div className="fixedExpensesCards">
                        {sortedItems.map((item) => (
                            <div key={item.id} className="fixedExpenseCard">
                                <div className="fixedExpenseCardHeader">
                                    <span className="fixedExpenseCardTitle">{item.description}</span>
                                    <span className="fixedExpenseCardAmount">{formatCurrency(item.amount_cents, 'MXN')}</span>
                                </div>
                                <div className="fixedExpenseCardMeta">
                                    <span className="fixedExpenseCardTag">
                                        {EXPENSE_TYPE_LABELS[item.expense_type] ?? item.expense_type}
                                    </span>
                                    <span className="fixedExpenseCardTag">
                                        {RECURRENCE_LABELS[item.recurrence_pattern] ?? item.recurrence_pattern}
                                    </span>
                                    <span className={`statusBadge ${item.is_active ? 'statusActive' : 'statusInactive'}`}>
                                        {item.is_active ? 'Activo' : 'Inactivo'}
                                    </span>
                                </div>
                                <div className="fixedExpenseCardActions">
                                    <button
                                        type="button"
                                        className="btn btnGhost btnSm"
                                        onClick={() => handleOpenCatalog(item)}
                                    >
                                        🔗 Asociar
                                    </button>
                                    <button
                                        type="button"
                                        className="btn btnGhost btnSm"
                                        onClick={() => handleEdit(item)}
                                    >
                                        ✏️ Editar
                                    </button>
                                    <button
                                        type="button"
                                        className="btn btnGhost btnSm"
                                        onClick={() => handleDelete(item)}
                                    >
                                        🗑️ Eliminar
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </section>

            {/* ─── KPI Panel ─── */}
            {householdId && items.length > 0 && (
                <div className="u-mt-3">
                    <button
                        type="button"
                        className="btn btnGhost"
                        onClick={() => setShowKPIPanel((prev) => !prev)}
                    >
                        {showKPIPanel ? 'Ocultar' : 'Mostrar'} análisis de suscripciones
                    </button>
                    {showKPIPanel && (
                        <div className="u-mt-2">
                            <SubscriptionKPIPanel householdId={householdId} />
                        </div>
                    )}
                </div>
            )}

            {/* ─── Catalog Selector ─── */}
            {catalogExpenseId && (
                <SubscriptionCatalogSelector
                    open={!!catalogExpenseId}
                    recurringExpenseId={catalogExpenseId}
                    linkedServiceIds={catalogLinkedIds}
                    onClose={handleCloseCatalog}
                    onSave={handleCatalogSave}
                />
            )}
        </motion.div>
    )
}
