import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { BottomNav } from '../BottomNav'

function renderNav(path = '/') {
  window.history.pushState({}, '', path)
  return render(
    <MemoryRouter initialEntries={[path]}>
      <BottomNav householdId="hh-1" />
    </MemoryRouter>,
  )
}

describe('BottomNav', () => {
  it('renders all five nav items with Spanish labels', () => {
    renderNav('/')

    expect(screen.getByText('Vista')).toBeInTheDocument()
    expect(screen.getByText('Movimientos')).toBeInTheDocument()
    expect(screen.getByText('Balances')).toBeInTheDocument()
    expect(screen.getByText('Plazos')).toBeInTheDocument()
    expect(screen.getByText('Config')).toBeInTheDocument()
  })

  it('renders Heroicons SVG icons with aria-hidden', () => {
    renderNav('/')

    const icons = document.querySelectorAll('.pillNavIcon[aria-hidden="true"]')
    expect(icons.length).toBe(5)
    icons.forEach((icon) => {
      expect(icon.tagName).toBe('svg')
    })
  })

  it('marks the active route with aria-current="page"', () => {
    renderNav('/expenses')

    const activeLink = screen.getByText('Movimientos').closest('a')
    expect(activeLink).toHaveAttribute('aria-current', 'page')
  })

  it('has no single-letter icons (V, M, B, P, C)', () => {
    renderNav('/')

    const iconElements = document.querySelectorAll('.pillNavLabel')
    iconElements.forEach((el) => {
      expect(el.textContent).not.toMatch(/^[VMNPC]$/)
    })
  })
})
