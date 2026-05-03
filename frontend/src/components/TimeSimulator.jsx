import { useState } from 'react'
import { advanceTime } from '../api'

export function TimeSimulator({ onAdvanced }) {
    const [days, setDays] = useState(1)
    const [busy, setBusy] = useState(false)

    if (window.location.hostname !== 'localhost' && !window.location.hostname.includes('127.0.0.1')) {
        return null
    }

    async function handleAdvance() {
        setBusy(true)
        try {
            await advanceTime(days)
            if (onAdvanced) onAdvanced()
            alert(`Avanzado ${days} días`)
        } catch (err) {
            alert(err.message)
        } finally {
            setBusy(false)
        }
    }

    return (
        <div style={{
            position: 'fixed',
            bottom: '80px',
            right: '20px',
            background: 'var(--card-bg)',
            padding: '10px',
            borderRadius: '8px',
            boxShadow: '0 4px 12px rgba(0,0,0,0.2)',
            zIndex: 1000,
            border: '1px solid var(--border)',
            display: 'flex',
            flexDirection: 'column',
            gap: '8px'
        }}>
            <p style={{ margin: 0, fontSize: '12px', fontWeight: 'bold' }}>Simulador (Dev)</p>
            <div style={{ display: 'flex', gap: '4px' }}>
                <input
                    type="number"
                    value={days}
                    onChange={(e) => setDays(e.target.value)}
                    style={{ width: '50px', padding: '4px' }}
                />
                <button
                    onClick={handleAdvance}
                    disabled={busy}
                    className="btn btnPrimary"
                    style={{ padding: '4px 8px', fontSize: '12px' }}
                >
                    +{days}d
                </button>
            </div>
        </div>
    )
}
