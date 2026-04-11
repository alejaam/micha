import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SettlementPanel } from '../SettlementPanel'

function renderPanelWithSettlement(settlement) {
  return render(
    <SettlementPanel
      settlement={settlement}
      settlementYear={2026}
      settlementMonth={4}
      onSettlementYearChange={() => {}}
      onSettlementMonthChange={() => {}}
      onRefresh={() => {}}
      onResetToCurrentMonth={() => {}}
      loadingSettlement={false}
      memberIndex={{ 'm-1': 'Ana', 'm-2': 'Luis' }}
      currency="MXN"
      selectedHousehold={{ created_at: '2025-01-01T00:00:00.000Z' }}
    />,
  )
}

describe('SettlementPanel semantic balance cards', () => {
  it('renders OWES state for negative balances', () => {
    renderPanelWithSettlement({
      effective_settlement_mode: 'equal',
      total_shared_cents: 40000,
      included_expense_count: 2,
      excluded_voucher_count: 0,
      transfers: [],
      members: [
        {
          member_id: 'm-1',
          salary_weight_bps: 5000,
          paid_cents: 10000,
          expected_share: 20000,
          net_balance_cents: -10000,
        },
      ],
    })

    expect(screen.getByText('[OWES]')).toBeInTheDocument()
    expect(screen.getByText('MX$100.00')).toBeInTheDocument()
  })

  it('renders RECEIVES state for positive balances', () => {
    renderPanelWithSettlement({
      effective_settlement_mode: 'equal',
      total_shared_cents: 40000,
      included_expense_count: 2,
      excluded_voucher_count: 0,
      transfers: [],
      members: [
        {
          member_id: 'm-2',
          salary_weight_bps: 5000,
          paid_cents: 30000,
          expected_share: 20000,
          net_balance_cents: 10000,
        },
      ],
    })

    expect(screen.getByText('[RECEIVES]')).toBeInTheDocument()
    expect(screen.getByText('MX$100.00')).toBeInTheDocument()
  })
})
