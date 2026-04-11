import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { formatCurrency } from '../utils'
import { EmptyState } from '../ui/EmptyState'

function moneyTick(value, currency) {
  return formatCurrency(Number(value ?? 0), currency)
}

export function HistorySection({
  closedPeriods,
  selectedPeriodKey,
  onSelectPeriod,
  comparisonSeries,
  memberBalanceTrend,
  completedMsi,
  selectedPeriodSnapshot,
  currency = 'MXN',
  isLoading = false,
  isProvisional = false,
  provisionalReason = '',
  onQuickAdd,
}) {
  const safeClosedPeriods = Array.isArray(closedPeriods) ? closedPeriods : []
  const safeComparisonSeries = Array.isArray(comparisonSeries) ? comparisonSeries : []
  const safeMemberBalanceTrend = Array.isArray(memberBalanceTrend) ? memberBalanceTrend : []
  const safeCompletedMsi = Array.isArray(completedMsi) ? completedMsi : []

  if (!isLoading && safeClosedPeriods.length === 0) {
    return (
      <section className="card" aria-label="History section">
        <div className="listHeader">
          <h2 className="listTitle">History</h2>
        </div>
        <EmptyState
          title="No closed periods yet"
          description="History becomes available after you move beyond the current period."
          ctaLabel="Add expense"
          onCta={onQuickAdd}
          icon="[H]"
          compact
        />
      </section>
    )
  }

  return (
    <section className="card historySection" aria-label="History section">
      <div className="listHeader">
        <h2 className="listTitle">History</h2>
        {isProvisional ? <span className="historyProvisionalTag">PROVISIONAL</span> : null}
      </div>

      {provisionalReason ? <p className="historyProvisionalNote">{provisionalReason}</p> : null}

      <div className="historyGrid">
        <article className="historyPanel" aria-label="Closed periods">
          <p className="historyPanelLabel">PERIODS</p>
          <ul className="historyPeriodList">
            {safeClosedPeriods.map((period) => (
              <li key={period.key}>
                <button
                  type="button"
                  className={`historyPeriodButton${selectedPeriodKey === period.key ? ' isSelected' : ''}`}
                  onClick={() => onSelectPeriod(period.key)}
                >
                  <span>{period.label}</span>
                  <span className="historyPeriodValue">{formatCurrency(period.totalCents, currency)}</span>
                </button>
              </li>
            ))}
          </ul>
        </article>

        <article className="historyPanel" aria-label="Historical totals comparison">
          <p className="historyPanelLabel">TOTALS</p>
          <div className="historyChartCanvas">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={safeComparisonSeries}>
                <XAxis dataKey="label" tick={{ fill: '#666666', fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis tickFormatter={(value) => moneyTick(value, currency)} tick={{ fill: '#666666', fontSize: 11 }} width={80} axisLine={false} tickLine={false} />
                <Tooltip formatter={(value) => moneyTick(value, currency)} />
                <Line type="monotone" dataKey="totalCents" stroke="#0A0A0A" strokeWidth={2} dot={{ r: 2 }} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </article>

        <article className="historyPanel" aria-label="Member balance trend">
          <p className="historyPanelLabel">BALANCES</p>
          <div className="historyChartCanvas">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={safeMemberBalanceTrend}>
                <XAxis dataKey="label" tick={{ fill: '#666666', fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis tickFormatter={(value) => moneyTick(value, currency)} tick={{ fill: '#666666', fontSize: 11 }} width={80} axisLine={false} tickLine={false} />
                <Tooltip formatter={(value) => moneyTick(value, currency)} />
                {Object.keys(safeMemberBalanceTrend[0] || {})
                  .filter((key) => !['key', 'label', 'source'].includes(key))
                  .slice(0, 3)
                  .map((memberName, index) => (
                    <Line
                      key={memberName}
                      type="monotone"
                      dataKey={memberName}
                      stroke={['#0A0A0A', '#4A4A4A', '#8A8A8A'][index % 3]}
                      strokeWidth={2}
                      dot={{ r: 2 }}
                    />
                  ))}
              </LineChart>
            </ResponsiveContainer>
          </div>
          {selectedPeriodSnapshot?.memberBalances?.length ? (
            <p className="historyPanelFootnote">
              Selected period members: {selectedPeriodSnapshot.memberBalances.length}
            </p>
          ) : null}
        </article>

        <article className="historyPanel" aria-label="Completed MSI history">
          <p className="historyPanelLabel">COMPLETED MSI</p>
          {safeCompletedMsi.length === 0 ? (
            <p className="historyPanelFootnote">No completed MSI expenses yet.</p>
          ) : (
            <ul className="historyMsiList">
              {safeCompletedMsi.map((item) => (
                <li key={item.id} className="historyMsiItem">
                  <span>{item.description}</span>
                  <span className="historyMsiFraction">{item.totalInstallments}/{item.totalInstallments}</span>
                </li>
              ))}
            </ul>
          )}
        </article>
      </div>
    </section>
  )
}
