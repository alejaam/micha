/**
 * FAB — Floating Action Button for adding a new expense.
 */
export function FAB({ onClick }) {
    return (
        <button
            type="button"
            className="fab"
            onClick={onClick}
            aria-label="Add new expense"
            title="Add expense"
        >
            <span aria-hidden>+</span>
        </button>
    )
}
