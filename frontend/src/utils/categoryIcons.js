/**
 * Category icon map — emoji icons for expense categories.
 * Used throughout the UI to add visual context.
 */

export const CATEGORY_ICONS = {
    rent: '🏠',
    auto: '🚗',
    streaming: '📺',
    food: '🍔',
    personal: '💄',
    savings: '💰',
    other: '📦',
}

/**
 * Get the icon for a category slug, with fallback.
 */
export function getCategoryIcon(categorySlug) {
    return CATEGORY_ICONS[categorySlug] || CATEGORY_ICONS.other
}