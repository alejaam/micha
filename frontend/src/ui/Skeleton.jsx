/**
 * Skeleton - Loading placeholder component
 * Shows animated placeholder while content loads
 */
export function Skeleton({ width = '100%', height = '20px', className = '' }) {
    return (
        <div
            className={`skeleton ${className}`}
            style={{ width, height }}
            aria-hidden="true"
        />
    )
}

export function SkeletonCard({ lines = 3 }) {
    return (
        <div className="card" aria-busy="true" aria-label="Loading...">
            <Skeleton width="60%" height="24px" className="mb-3" />
            {Array.from({ length: lines }).map((_, i) => (
                <Skeleton key={i} width={i === lines - 1 ? '80%' : '100%'} height="16px" className="mb-2" />
            ))}
        </div>
    )
}

export function SkeletonExpenseList({ count = 3 }) {
    return (
        <section className="card" aria-busy="true" aria-label="Loading expenses">
            <div className="listHeader">
                <Skeleton width="150px" height="24px" />
            </div>
            <ul className="expenseListUl" role="list">
                {Array.from({ length: count }).map((_, i) => (
                    <li key={i} className="expenseItem">
                        <div className="expenseBody">
                            <Skeleton width="70%" height="18px" className="mb-1" />
                            <Skeleton width="40%" height="14px" />
                        </div>
                        <div className="expenseRight">
                            <Skeleton width="80px" height="20px" />
                        </div>
                    </li>
                ))}
            </ul>
        </section>
    )
}
