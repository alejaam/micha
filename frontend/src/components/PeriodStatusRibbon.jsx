/**
 * PeriodStatusRibbon — compact inline chip for the current period lifecycle.
 * Renders as <span> for inline placement in header.
 *
 * Status → label mapping:
 *   open   → "Abierto"   (green)
 *   review → "Revisión"  (amber)
 *   closed → "Cerrado"   (gray)
 */
const CHIP_LABELS = {
    open: 'Abierto',
    review: 'Revisión',
    closed: 'Cerrado',
}

const CHIP_DESCRIPTIONS = {
    open: 'Periodo abierto — puedes registrar y editar gastos.',
    review: 'Periodo en revisión — las acciones de edición están bloqueadas temporalmente.',
    closed: 'Periodo cerrado — ya no se permiten cambios en gastos.',
}

export function PeriodStatusRibbon({ status = 'open' }) {
    const normalizedStatus = CHIP_LABELS[status] ? status : 'open'
    const label = CHIP_LABELS[normalizedStatus]
    const description = CHIP_DESCRIPTIONS[normalizedStatus]

    return (
        <span
            className={`periodChip periodChip-${normalizedStatus}`}
            role="status"
            aria-label={`Estado del periodo: ${description}`}
        >
            <span className="periodChipDot" aria-hidden>●</span>
            <span className="periodChipLabel">{label}</span>
        </span>
    )
}
