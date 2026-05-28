import { useCallback, useEffect, useState } from 'react';
import { getCurrentPeriod } from '../api';
import { useFocusRefetch } from './useFocusRefetch';

/**
 * Hook for managing the dashboard UI contextual state (ribbon status, active views).
 * Review/approval/consensus logic has been removed — periods only have open/closed status.
 */

export function useDashboardUxState(householdId) {
    const [currentPeriod, setCurrentPeriod] = useState(null);
    const [periodStatus, setPeriodStatus] = useState('open');
    const [isBottomSheetOpen, setIsBottomSheetOpen] = useState(false);
    const [isLoadingPeriod, setIsLoadingPeriod] = useState(false);

    const [selectedPeriodId, setSelectedPeriodId] = useState(null);

    const loadPeriod = useCallback(async () => {
        if (!householdId) return;
        try {
            setIsLoadingPeriod(true);
            const period = await getCurrentPeriod({ householdId });

            if (period) {
                setCurrentPeriod(period);
                setPeriodStatus(period.status || period.Status || 'open');
            } else {
                setCurrentPeriod(null);
                setPeriodStatus('open');
            }
        } catch (err) {
            console.error('Failed to load period:', err);
            setCurrentPeriod(null);
            setPeriodStatus('open');
        } finally {
            setIsLoadingPeriod(false);
        }
    }, [householdId])

    useEffect(() => {
        loadPeriod();
    }, [loadPeriod]);

    useFocusRefetch(loadPeriod)

    const openBottomSheet = () => setIsBottomSheetOpen(true);
    const closeBottomSheet = () => setIsBottomSheetOpen(false);

    return {
        currentPeriod,
        periodStatus,
        setPeriodStatus,
        isBottomSheetOpen,
        openBottomSheet,
        closeBottomSheet,
        loadPeriod,
        isLoadingPeriod,
        selectedPeriodId,
        setSelectedPeriodId,
    };
}
