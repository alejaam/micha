export function EmptyState({
  title,
  description,
  ctaLabel,
  onCta,
  icon = '[ ]',
  compact = false,
}) {
  return (
    <section className={`emptyStateBlock${compact ? ' emptyStateBlockCompact' : ''}`} aria-label={title}>
      <p className="emptyStateBlockIcon" aria-hidden>{icon}</p>
      <h3 className="emptyStateBlockTitle">{title}</h3>
      {description ? <p className="emptyStateBlockDescription">{description}</p> : null}
      {ctaLabel && onCta ? (
        <button type="button" className="btn btnGhost btnSm" onClick={onCta}>
          {ctaLabel}
        </button>
      ) : null}
    </section>
  )
}
