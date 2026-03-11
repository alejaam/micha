/**
 * Format an amount in cents as a localized currency string.
 * e.g. 4250, MXN -> "$42.50"
 */
export function formatCurrency(amountCents, currency = 'MXN') {
    return new Intl.NumberFormat(undefined, {
        style: 'currency',
        currency,
        minimumFractionDigits: 2,
    }).format(amountCents / 100)
}

/**
 * Backward-compatible alias used by existing components.
 */
export function formatDollars(amountCents) {
    return formatCurrency(amountCents, 'USD')
}

/**
 * Parse a dollar string entered by the user into integer cents.
 * Returns null if the value is not a valid non-negative number.
 * e.g. "12.50" → 1250 | "abc" → null
 */
export function dollarsToCents(value) {
    const cleaned = String(value).replace(/[^0-9.]/g, '')
    const parsed = parseFloat(cleaned)
    if (!Number.isFinite(parsed) || parsed < 0) return null
    return Math.round(parsed * 100)
}

/**
 * Format an ISO timestamp as a short relative date label.
 * Falls back to a locale date string when the date is far in the past.
 */
export function formatRelativeDate(isoString) {
    if (!isoString) return ''
    const date = new Date(isoString)
    const now = new Date()
    const diffMs = now - date
    const diffMin = Math.floor(diffMs / 60_000)
    const diffHr = Math.floor(diffMin / 60)
    const diffDay = Math.floor(diffHr / 24)

    if (diffMin < 1) return 'just now'
    if (diffMin < 60) return `${diffMin}m ago`
    if (diffHr < 24) return `${diffHr}h ago`
    if (diffDay < 7) return `${diffDay}d ago`
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
