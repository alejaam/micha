import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { CategoriesGrid } from '../DynamicChartsPanel'

describe('CategoriesGrid', () => {
  it('returns nothing when there are no categories', () => {
    const { container } = render(
      <CategoriesGrid categoryTotals={[]} />,
    )

    expect(container).toBeEmptyDOMElement()
  })

  it('renders category cards with uppercase labels and amounts', () => {
    render(
      <CategoriesGrid
        categoryTotals={[
          { key: 'food', label: 'Food', totalCents: 50000, percentage: 50 },
          { key: 'rent', label: 'Rent', totalCents: 30000, percentage: 30 },
          { key: 'transport', label: 'Transport', totalCents: 20000, percentage: 20 },
        ]}
        currency="MXN"
      />,
    )

    expect(screen.getByText('FOOD')).toBeInTheDocument()
    expect(screen.getByText('RENT')).toBeInTheDocument()
    expect(screen.getByText('TRANSPORT')).toBeInTheDocument()
    expect(screen.getByText(/^\$500\.00/)).toBeInTheDocument()
    expect(screen.getByText(/^\$300\.00/)).toBeInTheDocument()
    expect(screen.getByText(/^\$200\.00/)).toBeInTheDocument()
  })

  it('groups extra categories into "Otros" when more than 5 exist', () => {
    const categories = Array.from({ length: 7 }, (_, i) => ({
      key: `cat${i}`,
      label: `Category ${i}`,
      totalCents: (7 - i) * 1000,
      percentage: ((7 - i) * 1000) / 28000 * 100,
    }))

    render(<CategoriesGrid categoryTotals={categories} currency="MXN" />)

    expect(screen.getByText('OTROS')).toBeInTheDocument()
  })
})
