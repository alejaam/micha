import { motion } from 'framer-motion'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { formatCurrency } from '../utils'

const DONUT_NEUTRALS = ['#0A0A0A', '#2A2A2A', '#4A4A4A', '#6A6A6A', '#8A8A8A', '#AAAAAA']
const CHART_GRID = '#E8E8E8'
const CHART_TEXT = '#666666'

function moneyTick(value, currency) {
  return formatCurrency(value ?? 0, currency)
}

function chartTooltipValue(value, currency) {
  return formatCurrency(Number(value ?? 0), currency)
}

export function CategoryDonutChart({ data, currency }) {
  return (
    <div className="chartPanelCard" aria-label="Expenses by category">
      <p className="chartPanelLabel">CATEGORIES</p>
      <h3 className="chartPanelTitle">Expenses by category</h3>

      <div className="chartCanvas chartCanvasDonut">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Tooltip formatter={(value) => chartTooltipValue(value, currency)} />
            <Pie
              data={data}
              dataKey="totalCents"
              nameKey="label"
              innerRadius={56}
              outerRadius={90}
              paddingAngle={1}
              stroke="none"
            >
              {data.map((entry, index) => (
                <Cell key={`${entry.key}-${index}`} fill={DONUT_NEUTRALS[index % DONUT_NEUTRALS.length]} />
              ))}
            </Pie>
          </PieChart>
        </ResponsiveContainer>
      </div>

      <ul className="chartLegend" aria-label="Category totals summary">
        {data.slice(0, 4).map((entry) => (
          <li key={entry.key} className="chartLegendRow">
            <span className="chartLegendName">{entry.label}</span>
            <span className="chartLegendValue">{formatCurrency(entry.totalCents, currency)}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

export function MemberComparisonChart({ data, currency }) {
  return (
    <div className="chartPanelCard" aria-label="Member actual versus expected spending">
      <p className="chartPanelLabel">MEMBERS</p>
      <h3 className="chartPanelTitle">Actual vs expected spend</h3>

      <div className="chartCanvas">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} margin={{ top: 10, right: 8, left: 8, bottom: 8 }}>
            <CartesianGrid stroke={CHART_GRID} vertical={false} />
            <XAxis dataKey="memberName" tick={{ fill: CHART_TEXT, fontSize: 11 }} axisLine={false} tickLine={false} />
            <YAxis
              tickFormatter={(value) => moneyTick(value, currency)}
              tick={{ fill: CHART_TEXT, fontSize: 11 }}
              width={76}
              axisLine={false}
              tickLine={false}
            />
            <Tooltip formatter={(value) => chartTooltipValue(value, currency)} />
            <Bar dataKey="actualCents" fill="#0A0A0A" radius={[4, 4, 0, 0]} name="Actual" />
            <Bar dataKey="expectedCents" fill="#8A8A8A" radius={[4, 4, 0, 0]} name="Expected" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <p className="chartSummary" role="note">
        Dark bars represent actual paid amounts; gray bars represent expected share.
      </p>
    </div>
  )
}

export function SpendingTrendChart({ data, currency }) {
  return (
    <div className="chartPanelCard" aria-label="Household spending trend over time">
      <p className="chartPanelLabel">TREND</p>
      <h3 className="chartPanelTitle">Household spending over time</h3>

      <div className="chartCanvas">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data} margin={{ top: 10, right: 8, left: 8, bottom: 8 }}>
            <CartesianGrid stroke={CHART_GRID} vertical={false} />
            <XAxis dataKey="label" tick={{ fill: CHART_TEXT, fontSize: 11 }} axisLine={false} tickLine={false} />
            <YAxis
              tickFormatter={(value) => moneyTick(value, currency)}
              tick={{ fill: CHART_TEXT, fontSize: 11 }}
              width={76}
              axisLine={false}
              tickLine={false}
            />
            <Tooltip formatter={(value) => chartTooltipValue(value, currency)} />
            <Line
              type="monotone"
              dataKey="totalCents"
              stroke="#0A0A0A"
              strokeWidth={2}
              dot={{ r: 2, fill: '#0A0A0A' }}
              activeDot={{ r: 4 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <p className="chartSummary" role="note">
        Line values show monthly total household spending in {currency}.
      </p>
    </div>
  )
}

export function MsiProgressList({ data }) {
  return (
    <div className="chartPanelCard" aria-label="MSI progress">
      <p className="chartPanelLabel">MSI</p>
      <h3 className="chartPanelTitle">Installment progress</h3>

      <ul className="msiProgressList">
        {data.slice(0, 5).map((item) => (
          <li key={item.id} className="msiProgressRow">
            <div className="msiProgressHeader">
              <span className="msiProgressName">{item.description}</span>
              <span className="msiProgressFraction">
                {item.currentInstallment}/{item.totalInstallments}
              </span>
            </div>
            <div className="msiProgressTrack" aria-hidden>
              <span className="msiProgressFill" style={{ width: `${item.progressPercent}%` }} />
            </div>
            <p className="msiProgressMeta">{item.remainingInstallments} installments remaining</p>
          </li>
        ))}
      </ul>
    </div>
  )
}

export function DynamicChartsPanel({
  categoryTotals = [],
  memberActualVsExpected = [],
  msiProgress = [],
  spendingTrend = [],
  currency = 'MXN',
}) {
  const topCategory = categoryTotals[0]
  const topVariance = [...memberActualVsExpected]
    .sort((a, b) => Math.abs(b.deltaCents) - Math.abs(a.deltaCents))[0]
  const latestPoint = spendingTrend[spendingTrend.length - 1]

  const showDonut = categoryTotals.length > 0
  const showBar = memberActualVsExpected.length > 0
  const showTrend = spendingTrend.length > 0
  const showMsi = msiProgress.length > 0

  if (!showDonut && !showBar && !showTrend && !showMsi) {
    return null
  }

  return (
    <motion.section
      className="card dynamicChartsPanel"
      aria-label="Dynamic charts"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.18, ease: 'easeOut' }}
    >
      <div className="dynamicChartsHeader">
        <h2 className="sectionTitle">
          <span className="sectionTitleIcon" aria-hidden>◌</span>
          Dynamic charts
        </h2>
        <p className="dynamicChartsSubtitle">Live visual summary for this period.</p>
      </div>

      <div className="dynamicChartsGrid">
        {showDonut && <CategoryDonutChart data={categoryTotals} currency={currency} />}
        {showBar && <MemberComparisonChart data={memberActualVsExpected} currency={currency} />}
        {showTrend && <SpendingTrendChart data={spendingTrend} currency={currency} />}
        {showMsi && <MsiProgressList data={msiProgress} />}
      </div>

      <div className="dynamicChartsNarrative" aria-label="Charts textual summaries" tabIndex={0}>
        {topCategory && (
          <p>
            Top category: <strong>{topCategory.label}</strong> at{' '}
            <strong>{formatCurrency(topCategory.totalCents, currency)}</strong>
            {' '}({topCategory.percentage.toFixed(1)}%).
          </p>
        )}

        {topVariance && (
          <p>
            Largest variance: <strong>{topVariance.memberName}</strong>{' '}
            <span className={topVariance.deltaCents > 0 ? 'valueSemanticWarning' : 'valueSemanticSuccess'}>
              {topVariance.deltaCents > 0 ? 'over expected' : 'under expected'}{' '}
              by {formatCurrency(Math.abs(topVariance.deltaCents), currency)}
            </span>
            .
          </p>
        )}

        {latestPoint && (
          <p>
            Latest trend point ({latestPoint.label}):{' '}
            <strong>{formatCurrency(latestPoint.totalCents, currency)}</strong>.
          </p>
        )}

      </div>
    </motion.section>
  )
}
