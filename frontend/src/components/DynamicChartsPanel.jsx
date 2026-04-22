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

// Brand Blue Palette (matching CSS tokens)
const COLOR_BRAND_500 = '#3b82f6'
const COLOR_BRAND_600 = '#2563eb'
const COLOR_BRAND_700 = '#1d4ed8'
const COLOR_BRAND_400 = '#60a5fa'
const COLOR_BRAND_200 = '#bfdbfe'
const COLOR_BRAND_100 = '#dbeafe'

const DONUT_COLORS = [
    COLOR_BRAND_600,
    COLOR_BRAND_400,
    '#6366f1', // Indigo accent
    '#0ea5e9', // Sky accent
    COLOR_BRAND_200,
    '#94a3b8', // Slate muted
]
const CHART_GRID = '#e2e8f0' // Matching --color-border
const CHART_TEXT = '#64748b' // Matching --color-text-3

function moneyTick(value, currency) {
  return formatCurrency(value ?? 0, currency)
}

function chartTooltipValue(value, currency) {
  return formatCurrency(Number(value ?? 0), currency)
}

export function CategoryDonutChart({ data, currency }) {
  return (
    <div className="chartPanelCard" aria-label="Gastos por categoría">
      <p className="chartPanelLabel">CATEGORÍAS</p>
      <h3 className="chartPanelTitle">Distribución por categoría</h3>

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
                <Cell key={`${entry.key}-${index}`} fill={DONUT_COLORS[index % DONUT_COLORS.length]} />
              ))}
            </Pie>
          </PieChart>
        </ResponsiveContainer>
      </div>

      <ul className="chartLegend" aria-label="Resumen de totales por categoría">
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
    <div className="chartPanelCard" aria-label="Gasto real vs esperado por miembro">
      <p className="chartPanelLabel">MIEMBROS</p>
      <h3 className="chartPanelTitle">Gasto real vs cuota esperada</h3>

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
            <Bar dataKey="actualCents" fill={COLOR_BRAND_600} radius={[4, 4, 0, 0]} name="Real" />
            <Bar dataKey="expectedCents" fill={COLOR_BRAND_200} radius={[4, 4, 0, 0]} name="Esperado" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <p className="chartSummary" role="note">
        Barras azules: pagos reales; barras claras: cuota proporcional.
      </p>
    </div>
  )
}

export function SpendingTrendChart({ data, currency }) {
  return (
    <div className="chartPanelCard" aria-label="Tendencia de gasto del hogar">
      <p className="chartPanelLabel">TENDENCIA</p>
      <h3 className="chartPanelTitle">Historial de gastos mensuales</h3>

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
              stroke={COLOR_BRAND_600}
              strokeWidth={2}
              dot={{ r: 3, fill: COLOR_BRAND_600, strokeWidth: 0 }}
              activeDot={{ r: 5, strokeWidth: 0 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <p className="chartSummary" role="note">
        Total mensual de gastos compartidos en {currency}.
      </p>
    </div>
  )
}

export function MsiProgressList({ data }) {
  return (
    <div className="chartPanelCard" aria-label="Progreso de MSI">
      <p className="chartPanelLabel">PLAZOS</p>
      <h3 className="chartPanelTitle">Progreso de pagos a meses</h3>

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
              <span className="msiProgressFill" style={{ width: `${item.progressPercent}%`, background: COLOR_BRAND_600 }} />
            </div>
            <p className="msiProgressMeta">Quedan {item.remainingInstallments} mensualidades</p>
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
      aria-label="Gráficos dinámicos"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.18, ease: 'easeOut' }}
    >
      <div className="dynamicChartsHeader">
        <h2 className="sectionTitle">
          <span className="sectionTitleIcon" aria-hidden>◌</span>
          Gráficos dinámicos
        </h2>
        <p className="dynamicChartsSubtitle">Resumen visual en vivo de este periodo.</p>
      </div>

      <div className="dynamicChartsGrid">
        {showDonut && <CategoryDonutChart data={categoryTotals} currency={currency} />}
        {showBar && <MemberComparisonChart data={memberActualVsExpected} currency={currency} />}
        {showTrend && <SpendingTrendChart data={spendingTrend} currency={currency} />}
        {showMsi && <MsiProgressList data={msiProgress} />}
      </div>

      <div className="dynamicChartsNarrative" aria-label="Resumen textual de gráficos" tabIndex={0}>
        {topCategory && (
          <p>
            Categoría principal: <strong>{topCategory.label}</strong> con{' '}
            <strong>{formatCurrency(topCategory.totalCents, currency)}</strong>
            {' '}({topCategory.percentage.toFixed(1)}%).
          </p>
        )}

        {topVariance && (
          <p>
            Mayor varianza: <strong>{topVariance.memberName}</strong>{' '}
            <span className={topVariance.deltaCents > 0 ? 'valueSemanticWarning' : 'valueSemanticSuccess'}>
              {topVariance.deltaCents > 0 ? 'por encima de su cuota' : 'por debajo de su cuota'}{' '}
              por {formatCurrency(Math.abs(topVariance.deltaCents), currency)}
            </span>
            .
          </p>
        )}

        {latestPoint && (
          <p>
            Último punto de tendencia ({latestPoint.label}):{' '}
            <strong>{formatCurrency(latestPoint.totalCents, currency)}</strong>.
          </p>
        )}

      </div>
    </motion.section>
  )
}
