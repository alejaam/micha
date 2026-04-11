import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { BottomSheet } from '../BottomSheet'

describe('BottomSheet', () => {
  it('renders when open and closes on overlay click', () => {
    const onClose = vi.fn()
    render(
      <BottomSheet open title="Quick add" onClose={onClose}>
        <div>Body</div>
      </BottomSheet>,
    )

    expect(screen.getByRole('dialog', { name: /quick add/i })).toBeInTheDocument()
    fireEvent.click(document.querySelector('.bottomSheetOverlay'))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('closes on escape', () => {
    const onClose = vi.fn()
    render(
      <BottomSheet open title="Quick add" onClose={onClose}>
        <div>Body</div>
      </BottomSheet>,
    )

    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('keeps dialog semantics in panel and focuses close handle on open', () => {
    const onClose = vi.fn()
    render(
      <BottomSheet open title="Quick add" onClose={onClose}>
        <div>Body</div>
      </BottomSheet>,
    )

    const dialog = screen.getByRole('dialog', { name: /quick add/i })
    expect(dialog).toHaveAttribute('aria-modal', 'true')
    expect(dialog).toHaveAttribute('tabindex', '-1')
    expect(screen.getByRole('button', { name: /close panel/i })).toHaveFocus()
  })
})
