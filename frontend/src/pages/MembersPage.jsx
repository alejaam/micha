import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { useMembers } from '../hooks/useMembers'
import { Link } from 'react-router-dom'

export function MembersPage() {
    const { isAuthenticated, handleProtectedError } = useAuth()
    const { householdId, selectedHousehold } = useAppShell()

    const { members, loadingMembers } = useMembers({
        isAuthenticated,
        householdId,
        handleProtectedError,
    })

    return (
        <div className="pageWrapper" style={{ padding: '0 var(--sp-md)' }}>
            <header className="sectionHeader" style={{ marginBottom: 'var(--sp-wide)', borderBottom: '2px solid var(--color-text-1)', paddingBottom: '12px' }}>
                <span className="heroContext" style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', opacity: 0.6, textTransform: 'uppercase', letterSpacing: '0.1em' }}>
                    [ SYSTEM : HOUSEHOLD_MEMBERS ]
                </span>
                <h1 style={{ fontSize: '1.2rem', fontWeight: 'var(--fw-bold)', color: 'var(--color-text-1)', margin: '4px 0 0 0' }}>micha.members</h1>
            </header>

            <div className="membersGrid" style={{ display: 'grid', gap: 'var(--sp-medium)' }}>
                {loadingMembers ? (
                    <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}>LOADING_RECORDS...</p>
                ) : members.length === 0 ? (
                    <div className="card" style={{ border: '1px solid var(--color-border)', textAlign: 'center', padding: 'var(--sp-wide)' }}>
                        <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem', opacity: 0.6 }}>NO_MEMBERS_FOUND</p>
                    </div>
                ) : (
                    members.map(member => (
                        <div key={member.id} className="card" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', border: '1px solid var(--color-border)', padding: '16px' }}>
                            <div>
                                <h3 style={{ fontSize: '1rem', fontWeight: 'var(--fw-bold)', margin: 0 }}>{member.name}</h3>
                                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.7rem', opacity: 0.5, margin: '4px 0 0 0' }}>ID: {member.id.slice(0, 12)}</p>
                            </div>
                            <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', padding: '4px 8px', border: '1px solid var(--color-text-1)', borderRadius: 'var(--radius-pill)' }}>ACTIVE</span>
                        </div>
                    ))
                )}

                <Link
                    to="/members/new"
                    className="btn"
                    style={{ width: '100%', border: '2px solid var(--color-text-1)', fontFamily: 'var(--font-mono)', fontWeight: 'var(--fw-bold)', fontSize: '0.8rem', textAlign: 'center' }}
                >
                    + ADD_NEW_MEMBER
                </Link>
            </div>
        </div>
    )
}
