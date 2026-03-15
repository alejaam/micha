import { createContext, useContext } from 'react'

/**
 * AppShellContext distributes shared shell state (households, health, etc.)
 * to layouts and pages without prop-drilling through the router config.
 */
export const AppShellContext = createContext(null)

export function useAppShell() {
    const ctx = useContext(AppShellContext)
    if (!ctx) throw new Error('useAppShell must be used inside <AppShellProvider>')
    return ctx
}
