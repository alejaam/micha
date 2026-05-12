import { useEffect, useRef, useState } from 'react'

/**
 * UserMenu — avatar dropdown with household switching, health status, and logout.
 *
 * @param {{ email: string }} user  - User object from AuthContext
 * @param {Array}  households       - List of household objects
 * @param {string} householdId      - Currently selected household ID
 * @param {function} onHouseholdChange - Callback when selecting a household
 * @param {function} onLogout       - Callback for logout
 * @param {string} health           - Backend health status ('ok' | other)
 */
export function UserMenu({
    user,
    households = [],
    householdId,
    onHouseholdChange,
    onLogout,
    health,
}) {
    const [isOpen, setIsOpen] = useState(false)
    const containerRef = useRef(null)

    const initial = user?.email?.[0]?.toUpperCase() ?? '\u{1F464}'
    const isLive = health === 'ok'

    // Close on outside click
    useEffect(() => {
        if (!isOpen) return

        function handleClick(e) {
            if (containerRef.current && !containerRef.current.contains(e.target)) {
                setIsOpen(false)
            }
        }

        // Defer adding listener to avoid closing on the same click that opened it
        const timer = setTimeout(() => {
            document.addEventListener('click', handleClick)
        }, 0)

        return () => {
            clearTimeout(timer)
            document.removeEventListener('click', handleClick)
        }
    }, [isOpen])

    // Close on Escape key
    useEffect(() => {
        if (!isOpen) return

        function onKeyDown(e) {
            if (e.key === 'Escape') setIsOpen(false)
        }

        document.addEventListener('keydown', onKeyDown)
        return () => document.removeEventListener('keydown', onKeyDown)
    }, [isOpen])

    return (
        <div className="userMenu" ref={containerRef}>
            <button
                type="button"
                className="userMenuAvatar"
                onClick={() => setIsOpen((prev) => !prev)}
                aria-label="Menú de usuario"
                aria-haspopup="true"
                aria-expanded={isOpen}
            >
                {initial}
            </button>

            {isOpen && (
                <div className="userMenuDropdown" role="menu">
                    {/* User email */}
                    <div className="userMenuHeader">
                        {user?.email ?? 'Usuario'}
                    </div>

                    <div className="userMenuDivider" />

                    {/* Household list */}
                    {households.length > 0 ? (
                        <>
                            <div className="userMenuSectionLabel">
                                Hogares
                            </div>
                            {households.map((h) => (
                                <button
                                    key={h.id}
                                    type="button"
                                    className={
                                        'userMenuItem' +
                                        (h.id === householdId
                                            ? ' userMenuItemActive'
                                            : '')
                                    }
                                    role="menuitem"
                                    onClick={() => {
                                        onHouseholdChange(h.id)
                                        setIsOpen(false)
                                    }}
                                >
                                    {h.name}
                                    {h.id === householdId && (
                                        <span aria-hidden>&#10003;</span>
                                    )}
                                </button>
                            ))}
                        </>
                    ) : (
                        <div
                            className="userMenuEmpty"
                        >
                            No hay hogares
                        </div>
                    )}

                    <div className="userMenuDivider" />

                    {/* Footer: health + logout */}
                    <div className="userMenuFooter">
                        <span
                            className={
                                isLive
                                    ? 'pill pillOk userMenuPill'
                                    : 'pill pillOff userMenuPill'
                            }
                        >
                            {isLive ? 'live' : health}
                        </span>
                        <button
                            type="button"
                            className="btn btnGhostDanger btnSm"
                            onClick={onLogout}
                        >
                            Cerrar sesión
                        </button>
                    </div>
                </div>
            )}
        </div>
    )
}
