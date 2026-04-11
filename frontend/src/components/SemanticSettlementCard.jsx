import { formatCurrency } from '../utils'

/**
 * SemanticSettlementCard
 * Renders a member balance with semantic state:
 * - owes (negative)
 * - receives (positive)
 * - settled (zero)
 */
export function SemanticSettlementCard({
  memberName,
  netBalanceCents,
  currency = 'MXN',
}) {
  const amount = Number(netBalanceCents ?? 0)

  const state = amount < 0 ? 'owes' : amount > 0 ? 'receives' : 'settled'
  const stateLabel = state === 'owes' ? 'Owes' : state === 'receives' ? 'Receives' : 'Settled'
  const absoluteAmount = Math.abs(amount)

  return (
    <article
      className={`semanticSettlementCard semanticSettlementCard-${state}`}
      aria-label={`${memberName} ${stateLabel.toLowerCase()} ${formatCurrency(absoluteAmount, currency)}`}
    >
      <p className="semanticSettlementCardLabel">{memberName}</p>
      <div className="semanticSettlementCardValueRow">
        <span className="semanticSettlementCardState">[{stateLabel.toUpperCase()}]</span>
        <span className="semanticSettlementCardAmount">
          {formatCurrency(absoluteAmount, currency)}
        </span>
      </div>
    </article>
  )
}
