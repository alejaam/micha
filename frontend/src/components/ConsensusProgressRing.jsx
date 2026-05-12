import { motion } from 'framer-motion'

export function ConsensusProgressRing({ approved = 0, total = 0, label = 'Consensus', source = 'derived' }) {
  const safeTotal = Math.max(0, Number(total) || 0)
  const safeApproved = Math.min(safeTotal, Math.max(0, Number(approved) || 0))
  const percent = safeTotal > 0 ? Math.round((safeApproved / safeTotal) * 100) : 0

  const size = 96
  const strokeWidth = 8
  const radius = (size - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (percent / 100) * circumference

  return (
    <section className="consensusRing" aria-label={`${label} ${percent}%`}>
      <p className="consensusRingLabel">{label}</p>

      <div className="consensusRingCanvas" role="img" aria-label={`Consensus progress ${percent}%`}>
        <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke="var(--color-border)"
            strokeWidth={strokeWidth}
          />
          <motion.circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke={percent >= 100 ? 'var(--color-green-600)' : percent >= 50 ? 'var(--color-amber-400)' : 'var(--color-text-1)'}
            strokeWidth={strokeWidth}
            strokeLinecap="butt"
            strokeDasharray={circumference}
            initial={{ strokeDashoffset: circumference }}
            animate={{ strokeDashoffset: offset }}
            transition={{ duration: 0.2, ease: 'easeOut' }}
            transform={`rotate(-90 ${size / 2} ${size / 2})`}
          />
        </svg>
        <span className="consensusRingValue">{percent}%</span>
      </div>

      <p className="consensusRingMeta">
        {safeApproved}/{safeTotal} cuotas
        {source === 'mock' ? ' · provisional' : ''}
      </p>
    </section>
  )
}
