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
  it('renders all five nav items with emoji and Spanish labels', () => {
    renderNav('/')

    const resumen = screen.getByText('Resumen')
    const movimientos = screen.getByText('Movimientos')
    const balances = screen.getByText('Balances')
    const plazos = screen.getByText('Plazos')
    const reglas = screen.getByText('Reglas')

    expect(resumen).toBeInTheDocument()
    expect(movimientos).toBeInTheDocument()
    expect(balances).toBeInTheDocument()
    expect(plazos).toBeInTheDocument()
    expect(reglas).toBeInTheDocument()
  })

  it('shows emoji icons with aria-hidden', () => {
    renderNav('/')

    const icons = document.querySelectorAll('.bottomNavIcon[aria-hidden="true"]')
    expect(icons.length).toBe(5)
    expect(icons[0].textContent).toBe('📊')
    expect(icons[1].textContent).toBe('💸')
    expect(icons[2].textContent).toBe('⚖️')
    expect(icons[3].textContent).toBe('📅')
    expect(icons[4].textContent).toBe('⚙️')
  })

  it('marks the active route with aria-current="page"', () => {
    renderNav('/expenses')

    const activeLink = screen.getByText('Movimientos').closest('a')
    expect(activeLink).toHaveAttribute('aria-current', 'page')
  })

  it('has no single-letter icons (O, M, B, P, R)', () => {
    renderNav('/')

    const iconElements = document.querySelectorAll('.bottomNavIcon')
    iconElements.forEach((el) => {
      expect(el.textContent).not.toMatch(/^[OMBPR]$/)
    })
  })
})
