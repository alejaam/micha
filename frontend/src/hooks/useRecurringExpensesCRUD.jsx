import { useCallback, useEffect, useState } from 'react'
import {
    createRecurringExpense,
    deleteRecurringExpense,
    listRecurringExpenses,
    updateRecurringExpense,
} from '../api'
import { useAppShell } from '../context/AppShellContext'

/**
 * useRecurringExpensesCRUD — focused hook for CRUD mutations on recurring expenses.
 *
 * @param {Object} opts
 * @param {string} opts.householdId
 * @returns {{ items, loading, error, loadItems, createItem, updateItem, deleteItem, editingId, setEditingId }}
 */
export function useRecurringExpensesCRUD({ householdId }) {
    const { handleProtectedError } = useAppShell()
    const [items, setItems] = useState([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)
    const [editingId, setEditingId] = useState(null)

    const loadItems = useCallback(async () => {
        if (!householdId) return
        setLoading(true)
        setError(null)
        try {
            const data = await listRecurringExpenses({ householdId })
            setItems(Array.isArray(data) ? data : [])
        } catch (err) {
            handleProtectedError(err)
            setError(err.message ?? 'Error al cargar gastos fijos')
        } finally {
            setLoading(false)
        }
    }, [householdId, handleProtectedError])

    useEffect(() => {
        loadItems()
    }, [loadItems])

    const createItem = useCallback(async (payload) => {
        setError(null)
        try {
            const result = await createRecurringExpense(payload)
            // Optimistic: reload full list
            await loadItems()
            return result
        } catch (err) {
            handleProtectedError(err)
            setError(err.message ?? 'Error al crear gasto fijo')
            return null
        }
    }, [loadItems, handleProtectedError])

    const updateItem = useCallback(async (recurringExpenseId, payload) => {
        setError(null)
        try {
            await updateRecurringExpense({ recurringExpenseId, ...payload })
            await loadItems()
            setEditingId(null)
            return true
        } catch (err) {
            handleProtectedError(err)
            setError(err.message ?? 'Error al actualizar gasto fijo')
            return false
        }
    }, [loadItems, handleProtectedError])

    const deleteItem = useCallback(async (recurringExpenseId) => {
        setError(null)
        try {
            await deleteRecurringExpense({ recurringExpenseId })
            // Optimistic removal
            setItems((prev) => prev.filter((item) => item.id !== recurringExpenseId))
            return true
        } catch (err) {
            handleProtectedError(err)
            setError(err.message ?? 'Error al eliminar gasto fijo')
            return false
        }
    }, [handleProtectedError])

    return {
        items,
        loading,
        error,
        loadItems,
        createItem,
        updateItem,
        deleteItem,
        editingId,
        setEditingId,
    }
}
