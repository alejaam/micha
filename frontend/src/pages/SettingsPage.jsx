import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { Banner } from '../ui/Banner'
import { useState } from 'react'

export function SettingsPage() {
    const { logout } = useAuth()
    const { selectedHousehold, health, setHouseholdId, households, handleReload } = useAppShell()
    const [message, setMessage] = useState('')

    const handleLogout = () => {
        logout()
    }

    return (
        <div className="pageWrapper" style={{ padding: '0 var(--sp-md)' }}>
            {message && <Banner type="ok" onDismiss={() => setMessage('')}>{message}</Banner>}
            
            <header className="sectionHeader" style={{ marginBottom: 'var(--sp-wide)', borderBottom: '2px solid var(--color-text-1)', paddingBottom: '12px' }}>
                <span className="heroContext" style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', opacity: 0.6, textTransform: 'uppercase', letterSpacing: '0.1em' }}>
                    [ SYSTEM : CONFIGURATION ]
                </span>
                <h1 style={{ fontSize: '1.2rem', fontWeight: 'var(--fw-bold)', color: 'var(--color-text-1)', margin: '4px 0 0 0' }}>micha.settings</h1>
            </header>

            <div className="settingsStack" style={{ display: 'grid', gap: 'var(--sp-wide)' }}>
                {/* Household Switcher Section */}
                {households.length > 1 && (
                    <section className="settingsGroup">
                        <h2 style={{ fontFamily: 'var(--font-mono)', fontSize: '0.85rem', textTransform: 'uppercase', marginBottom: '12px' }}>Context: Switch_Household</h2>
                        <div style={{ display: 'grid', gap: '8px' }}>
                            {households.map(h => (
                                <button 
                                    key={h.id}
                                    onClick={() => { setHouseholdId(h.id); setMessage(`Context switched to ${h.name}`) }}
                                    className="card"
                                    style={{ 
                                        textAlign: 'left', 
                                        padding: '12px', 
                                        border: h.id === selectedHousehold?.id ? '2px solid var(--color-text-1)' : '1px solid var(--color-border-soft)',
                                        background: h.id === selectedHousehold?.id ? 'var(--color-surface-2)' : 'transparent',
                                        cursor: 'pointer'
                                    }}
                                >
                                    <span style={{ fontWeight: 'var(--fw-bold)' }}>{h.name}</span>
                                    {h.id === selectedHousehold?.id && <span style={{ marginLeft: '8px', fontSize: '0.65rem', fontFamily: 'var(--font-mono)', background: 'var(--color-text-1)', color: 'var(--color-bg)', padding: '2px 4px' }}>CURRENT</span>}
                                </button>
                            ))}
                        </div>
                    </section>
                )}

                {/* System Info Section */}
                <section className="settingsGroup">
                    <h2 style={{ fontFamily: 'var(--font-mono)', fontSize: '0.85rem', textTransform: 'uppercase', marginBottom: '12px' }}>System: Diagnostics</h2>
                    <div className="card" style={{ border: '1px solid var(--color-border-soft)', padding: '16px', fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                            <span style={{ opacity: 0.5 }}>BACKEND_HEALTH:</span>
                            <span style={{ color: health === 'ok' ? 'var(--color-success)' : 'var(--color-error)' }}>{health.toUpperCase()}</span>
                        </div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                            <span style={{ opacity: 0.5 }}>ACTIVE_HOUSEHOLD_ID:</span>
                            <span>{selectedHousehold?.id?.slice(0, 16)}...</span>
                        </div>
                         <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                            <span style={{ opacity: 0.5 }}>APP_VERSION:</span>
                            <span>0.0.1_STABLE</span>
                        </div>
                    </div>
                </section>

                {/* Account Section */}
                <section className="settingsGroup" style={{ borderTop: '1px solid var(--color-border-soft)', paddingTop: 'var(--sp-wide)' }}>
                    <button 
                        onClick={handleLogout}
                        className="btn"
                        style={{ width: '100%', border: '2px solid var(--color-error)', color: 'var(--color-error)', fontFamily: 'var(--font-mono)', fontWeight: 'var(--fw-bold)', fontSize: '0.8rem' }}
                    >
                        [ SIGN_OUT_SYSTEM ]
                    </button>
                </section>
            </div>
        </div>
    )
}
