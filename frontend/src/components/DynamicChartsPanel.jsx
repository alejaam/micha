import { formatCurrency } from '../utils'

/**
 * CategoriesGrid — redesigned DynamicChartsPanel. Shows top 5 spending
 * categories as compact cards with an uppercase label, total, and thin
 * progress bar proportional to total spending.
 *
 * @param {Array}  categoryTotals - [{ key, label, totalCents, percentage }]
 * @param {string} currency
 */
export function CategoriesGrid({ categoryTotals = [], currency = 'MXN' }) {
    if (!Array.isArray(categoryTotals) || categoryTotals.length === 0) {
        return null
    }

    // Top 5 categories + "Otros" bucket if more exist
    const top = categoryTotals.slice(0, 5)
    const others = categoryTotals.slice(5)

    const hasOthers = others.length > 0
    const othersTotal = hasOthers
        ? others.reduce((sum, c) => sum + c.totalCents, 0)
        : 0

    const grandTotal = categoryTotals.reduce((sum, c) => sum + c.totalCents, 0)

    const displayItems = hasOthers
        ? [
              ...top,
              {
                  key: '__otros__',
                  label: 'Otros',
                  totalCents: othersTotal,
                  percentage: grandTotal > 0 ? (othersTotal / grandTotal) * 100 : 0,
              },
          ]
        : top

    return (
        <section className="card" aria-label="Categorías de gasto">
            <h2 className="sectionTitle">
                <span className="sectionTitleIcon" aria-hidden>◈</span>
                Categorías
                {categoryTotals.length > 0 && (
                    <span className="sectionBadge">{categoryTotals.length} categorías</span>
                )}
            </h2>
            <div className="categoriesGrid">
                {displayItems.map((cat) => (
                    <div key={cat.key} className="categoryCard">
                        <div className="categoryCardName">{cat.label.toUpperCase()}</div>
                        <div className="categoryCardAmount">
                            {formatCurrency(cat.totalCents, currency)}
                        </div>
                        <div className="categoryCardBar" aria-hidden>
                            <div
                                className="categoryCardBarFill"
                                style={{ width: `${Math.min(cat.percentage, 100)}%` }}
                            />
                        </div>
                    </div>
                ))}
            </div>
        </section>
    )
}
