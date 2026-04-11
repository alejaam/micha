/**
 * FAB — Floating Action Button for adding a new expense.
 */
export function FAB({ onClick, disabled = false }) {
    return (
        <button
            type="button"
            className={`fab${disabled ? ' fabDisabled' : ''}`}
            onClick={onClick}
            disabled={disabled}
            aria-label="Add new expense"
            title="Add expense"
        >
            <span aria-hidden>+</span>
        </button>
    )
}
