import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ConsensusProgressRing } from '../ConsensusProgressRing'

describe('ConsensusProgressRing', () => {
  it('shows percentage and approvals summary', () => {
    render(<ConsensusProgressRing approved={1} total={2} />)

    expect(screen.getByText('50%')).toBeInTheDocument()
    expect(screen.getByText(/1\/2 approvals/i)).toBeInTheDocument()
  })
})
