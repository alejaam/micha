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

export async function createRecurringExpense({
    householdId,
    paidByMemberId = '',
    isAgnostic = false,
    amountCents,
    description,
    category = 'other',
    expenseType = 'fixed',
    recurrencePattern = 'monthly',
    startDate,
    endDate = null,
}) {
    const body = {
        household_id: householdId,
        paid_by_member_id: paidByMemberId,
        is_agnostic: isAgnostic,
        amount_cents: amountCents,
        description,
        category_id: category,
        expense_type: expenseType,
        recurrence_pattern: recurrencePattern,
        start_date: startDate,
    }
    if (endDate) body.end_date = endDate

    const response = await fetch('/v1/recurring-expenses', {
        method: 'POST',
        headers: buildProtectedHeaders(),
        body: JSON.stringify(body),
    })
    return parseResponse(response)
}

export async function listRecurringExpenses({ householdId, limit = 100, offset = 0 }) {
    const params = new URLSearchParams({
        household_id: householdId,
        limit: String(limit),
        offset: String(offset),
    })
    const response = await fetch(`/v1/recurring-expenses?${params.toString()}`, {
        headers: buildProtectedHeaders(),
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

export async function getRemainingSalary({ householdId, memberId, from, to }) {
    const params = new URLSearchParams()
    if (from) params.append('from', from)
    if (to) params.append('to', to)

    const response = await fetch(`/v1/households/${householdId}/members/${memberId}/remaining-salary?${params.toString()}`, {
        headers: buildProtectedHeaders(),
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

export async function getMe() {
    const response = await fetch('/v1/auth/me', {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

// ─── Period Management ────────────────────────────────────────────────────────

export async function getCurrentPeriod({ householdId }) {
    const response = await fetch(`/v1/households/${householdId}/periods/current`, {
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function initializePeriod({ householdId }) {
    const response = await fetch(`/v1/households/${householdId}/periods/initialize`, {
        method: 'POST',
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function transitionPeriodToReview({ householdId, periodId }) {
    const response = await fetch(`/v1/households/${householdId}/periods/${periodId}/review`, {
        method: 'POST',
        headers: buildProtectedHeaders(),
    })
    return parseResponse(response)
}

export async function approvePeriod({ householdId, periodId, status, comment = '' }) {
    const response = await fetch(`/v1/households/${householdId}/periods/${periodId}/approve`, {
        method: 'POST',
        headers: buildProtectedHeaders(),
        body: JSON.stringify({ status, comment }),
    })
    return parseResponse(response)
}

export async function closePeriod({ householdId, periodId, force = false }) {
    const response = await fetch(`/v1/households/${householdId}/periods/${periodId}/close`, {
        method: 'POST',
        headers: buildProtectedHeaders(),
        body: JSON.stringify({ force }),
    })
    return parseResponse(response)
}
