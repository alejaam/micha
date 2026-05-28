import { Navigate, Outlet } from 'react-router-dom'
import { AppHeader } from '../components/AppHeader'
import { BottomNav } from '../components/BottomNav'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { HouseholdDataProvider } from '../hooks/useHouseholdData'

/**
 * AppLayout — wraps protected routes.
 * Redirects unauthenticated users to /login.
 * Provides the global header, reading shared state from AppShellContext.
 */
export function AppLayout() {
    const { isAuthenticated, logout } = useAuth()
    const {
        health,
        householdId,
        households,
        setHouseholdId,
        handleReload,
        loadingHouseholds,
        periodStatus,
        currentPeriod,
        selectedPeriodId,
        setSelectedPeriodId,
    } = useAppShell()

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />
    }

    // No households yet → redirect to onboarding
    if (!loadingHouseholds && households.length === 0) {
        return <Navigate to="/onboarding/household" replace />
    }

    return (
        <HouseholdDataProvider>
            <div className="page">
                <AppHeader
                    health={health}
                    householdId={householdId}
                    households={households}
                    onHouseholdChange={setHouseholdId}
                    onReload={handleReload}
                    onLogout={logout}
                    isLoading={loadingHouseholds}
                    periodStatus={periodStatus}
                    currentPeriod={currentPeriod}
                    selectedPeriodId={selectedPeriodId}
                    onSelectPeriod={setSelectedPeriodId}
                />
                <Outlet />
                <BottomNav householdId={householdId} />
            </div>
        </HouseholdDataProvider>
    )
}
