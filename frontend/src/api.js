const JSON_HEADERS = {
    'Content-Type': 'application/json',
}

async function parseResponse(response) {
    if (response.status === 204) {
        return null
    }

    const payload = await response.json().catch(() => ({}))

    if (!response.ok) {
        const message = payload?.error?.message ?? `request failed with status ${response.status}`
        throw new Error(message)
    }

    return payload?.data
}

export async function getHealth() {
    const response = await fetch('/health')

    if (!response.ok) {
        throw new Error('backend no disponible')
    }

    return response.text()
}

export async function listExpenses({ householdId, limit = 20, offset = 0 }) {
    const params = new URLSearchParams({
        household_id: householdId,
        limit: String(limit),
        offset: String(offset),
    })

    const response = await fetch(`/v1/expenses?${params.toString()}`)
    return parseResponse(response)
}

export async function createExpense({ householdId, paidByMemberId, amountCents, description, isShared = true, currency = 'MXN', paymentMethod = 'cash' }) {
    const response = await fetch('/v1/expenses', {
        method: 'POST',
        headers: JSON_HEADERS,
        body: JSON.stringify({
            household_id: householdId,
            paid_by_member_id: paidByMemberId,
            amount_cents: amountCents,
            description,
            is_shared: isShared,
            currency,
            payment_method: paymentMethod,
        }),
    })

    return parseResponse(response)
}

export async function getSettlement({ householdId, year, month }) {
    const params = new URLSearchParams({
        year: String(year),
        month: String(month),
    })

    const response = await fetch(`/v1/households/${householdId}/settlement?${params.toString()}`)
    return parseResponse(response)
}

export async function listMembers({ householdId, limit = 100, offset = 0 }) {
    const params = new URLSearchParams({
        limit: String(limit),
        offset: String(offset),
    })

    const response = await fetch(`/v1/households/${householdId}/members?${params.toString()}`)
    return parseResponse(response)
}

export async function patchExpense({ id, amountCents, description }) {
    const body = {}

    if (typeof amountCents === 'number') {
        body.amount_cents = amountCents
    }

    if (typeof description === 'string') {
        body.description = description
    }

    const response = await fetch(`/v1/expenses/${id}`, {
        method: 'PATCH',
        headers: JSON_HEADERS,
        body: JSON.stringify(body),
    })

    return parseResponse(response)
}

export async function deleteExpense(id) {
    const response = await fetch(`/v1/expenses/${id}`, {
        method: 'DELETE',
    })

    return parseResponse(response)
}
