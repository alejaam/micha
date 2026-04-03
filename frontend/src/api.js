const JSON_HEADERS = {
    'Content-Type': 'application/json',
}

let authToken =
    typeof window !== 'undefined'
        ? (window.localStorage.getItem('micha_token') ?? '').trim()
        : ''

export function setAuthToken(token) {
    authToken = typeof token === 'string' ? token.trim() : ''
}

function buildProtectedHeaders() {
    if (!authToken) {
        return JSON_HEADERS
    }

    return {
        ...JSON_HEADERS,
        Authorization: `Bearer ${authToken}`,
    }
}

async function parseResponse(response) {
    if (response.status === 204) {
        return null
    }

    const payload = await response.json().catch(() => ({}))

    if (!response.ok) {
        const message = payload?.error?.message ?? `request failed with status ${response.status}`
        const error = new Error(message)
        error.code = payload?.error?.code ?? ''
        throw error
    }

    return payload?.data
}

export async function getHealth() {
    const response = await fetch('/health')

    if (!response.ok) {
        throw new Error('backend unavailable')
    }

    return response.text()
}

export async function listExpenses({ householdId, limit = 20, offset = 0 }) {
    const params = new URLSearchParams({
        household_id: householdId,
        limit: String(limit),
        offset: String(offset),
    })

    const response = await fetch(`/v1/expenses?${params.toString()}`, {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function listHouseholds({ limit = 100, offset = 0 } = {}) {
    const params = new URLSearchParams({
        limit: String(limit),
        offset: String(offset),
    })

    const response = await fetch(`/v1/households?${params.toString()}`, {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function getHousehold({ householdId }) {
    const response = await fetch(`/v1/households/${householdId}`, {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function updateHousehold({ householdId, name, settlementMode, currency }) {
    const response = await fetch(`/v1/households/${householdId}`, {
        method: 'PUT',
        headers: buildProtectedHeaders(),
        body: JSON.stringify({
            name,
            settlement_mode: settlementMode,
            currency,
        }),
    })

    return parseResponse(response)
}

export async function updateSplitConfig({ householdId, splits }) {
    const response = await fetch(`/v1/households/${householdId}/split-config`, {
        method: 'PUT',
        headers: buildProtectedHeaders(),
        body: JSON.stringify({
            splits: (splits || []).map((s) => ({
                member_id: s.memberId,
                percentage: s.percentage,
            })),
        }),
    })

    return parseResponse(response)
}

export async function registerUser({ email, password }) {
    const response = await fetch('/v1/auth/register', {
        method: 'POST',
        headers: JSON_HEADERS,
        body: JSON.stringify({ email, password }),
    })

    return parseResponse(response)
}

export async function loginUser({ email, password }) {
    const response = await fetch('/v1/auth/login', {
        method: 'POST',
        headers: JSON_HEADERS,
        body: JSON.stringify({ email, password }),
    })

    return parseResponse(response)
}

export async function createHousehold({ name, settlementMode = 'equal', currency = 'MXN' }) {
    const response = await fetch('/v1/households', {
        method: 'POST',
        headers: buildProtectedHeaders(),
        body: JSON.stringify({
            name,
            settlement_mode: settlementMode,
            currency,
        }),
    })

    return parseResponse(response)
}

export async function createMember({ householdId, name, email, monthlySalaryCents = 0 }) {
    const response = await fetch(`/v1/households/${householdId}/members`, {
        method: 'POST',
        headers: buildProtectedHeaders(),
        body: JSON.stringify({
            name,
            email,
            monthly_salary_cents: monthlySalaryCents,
        }),
    })

    return parseResponse(response)
}

export async function createExpense({ householdId, paidByMemberId, amountCents, description, isShared = true, currency = 'MXN', paymentMethod = 'cash', expenseType = 'variable', cardId = '', cardName = '', category = 'other', totalInstallments = 0 }) {
    const response = await fetch('/v1/expenses', {
        method: 'POST',
        headers: buildProtectedHeaders(),
        body: JSON.stringify({
            household_id: householdId,
            paid_by_member_id: paidByMemberId,
            amount_cents: amountCents,
            description,
            is_shared: isShared,
            currency,
            payment_method: paymentMethod,
            expense_type: expenseType,
            card_id: cardId,
            card_name: cardName,
            category,
            total_installments: totalInstallments,
        }),
    })

    return parseResponse(response)
}

export async function getSettlement({ householdId, year, month }) {
    const params = new URLSearchParams({
        year: String(year),
        month: String(month),
    })

    const response = await fetch(`/v1/households/${householdId}/settlement?${params.toString()}`, {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function listMembers({ householdId, limit = 100, offset = 0 }) {
    const params = new URLSearchParams({
        limit: String(limit),
        offset: String(offset),
    })

    const response = await fetch(`/v1/households/${householdId}/members?${params.toString()}`, {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function listCategories({ householdId }) {
    const response = await fetch(`/v1/households/${householdId}/categories`, {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function listCards({ householdId }) {
    const response = await fetch(`/v1/households/${householdId}/cards`, {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function createCard({ householdId, bankName, cardName, cutoffDay }) {
    const response = await fetch(`/v1/households/${householdId}/cards`, {
        method: 'POST',
        headers: buildProtectedHeaders(),
        body: JSON.stringify({
            bank_name: bankName,
            card_name: cardName,
            cutoff_day: cutoffDay,
        }),
    })

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
        headers: buildProtectedHeaders(),
        body: JSON.stringify(body),
    })

    return parseResponse(response)
}

export async function deleteExpense(id) {
    const response = await fetch(`/v1/expenses/${id}`, {
        method: 'DELETE',
        headers: buildProtectedHeaders(),
    })

    return parseResponse(response)
}

export async function generateRecurringExpenses({ householdId, asOfDate = null }) {
    const body = {
        household_id: householdId,
    }

    if (asOfDate) {
        body.as_of_date = asOfDate
    }

    const response = await fetch('/v1/recurring-expenses/generate', {
        method: 'POST',
        headers: buildProtectedHeaders(),
        body: JSON.stringify(body),
    })

    return parseResponse(response)
}

export async function getMe() {
    const response = await fetch('/v1/auth/me', {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}
