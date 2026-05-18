import { useCallback, useEffect, useState } from 'react';
import { getCurrentPeriod, getPeriodConsensus } from '../api';

/**
 * Hook for managing the dashboard UI contextual state (ribbon status, active views).
 */

export function buildConsensusState({ approved = 0, total = 0, source = 'derived' } = {}) {
    const safeTotal = Math.max(0, Number(total) || 0)
    const safeApproved = Math.min(safeTotal, Math.max(0, Number(approved) || 0))
    const percent = safeTotal > 0 ? Math.round((safeApproved / safeTotal) * 100) : 0

    return {
        approved: safeApproved,
        total: safeTotal,
        percent,
        source,
    }
}

export function useDashboardUxState(householdId) {
    const [currentPeriod, setCurrentPeriod] = useState(null);
    const [periodStatus, setPeriodStatus] = useState('open');
    const [isBottomSheetOpen, setIsBottomSheetOpen] = useState(false);
    const [isLoadingPeriod, setIsLoadingPeriod] = useState(false);

    // C9: Selected period state
    const [selectedPeriodId, setSelectedPeriodId] = useState(null);

    // C7: Real consensus state
    const [consensus, setConsensus] = useState({ approved: 0, total: 0, percent: 0, source: 'pending' });
    const [consensusLoading, setConsensusLoading] = useState(false);

    const loadPeriod = useCallback(async () => {
        if (!householdId) return;
        try {
            setIsLoadingPeriod(true);
            const period = await getCurrentPeriod({ householdId });
            
            if (period) {
                setCurrentPeriod(period);
                setPeriodStatus(period.status || period.Status || 'open');
            } else {
                // If API returns null data, it means no open period exists.
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

    // Load consensus when household and period are available
    const loadConsensus = useCallback(async (periodId) => {
        if (!householdId || !periodId) {
            setConsensus({ approved: 0, total: 0, percent: 0, source: 'pending' });
            return;
        }
        try {
            setConsensusLoading(true);
            const data = await getPeriodConsensus({ householdId, periodId });
            if (data) {
                setConsensus(buildConsensusState({
                    approved: data.approved ?? data.approved_count ?? 0,
                    total: data.total ?? data.total_members ?? 0,
                    source: 'api',
                }));
            } else {
                setConsensus({ approved: 0, total: 0, percent: 0, source: 'empty' });
            }
        } catch (err) {
            console.error('Failed to load consensus:', err);
            setConsensus({ approved: 0, total: 0, percent: 0, source: 'error' });
        } finally {
            setConsensusLoading(false);
        }
    }, [householdId]);

    useEffect(() => {
        loadPeriod();
    }, [loadPeriod]);

    // Reload consensus when period changes
    useEffect(() => {
        const targetId = selectedPeriodId || currentPeriod?.id;
        if (targetId) {
            loadConsensus(targetId);
        }
    }, [currentPeriod?.id, selectedPeriodId, loadConsensus]);

    const openBottomSheet = () => setIsBottomSheetOpen(true);
    const closeBottomSheet = () => setIsBottomSheetOpen(false);

    // Business rule: lock mutations during 'review'
    const statusForLock = periodStatus === 'review' || periodStatus === 'closed' ? periodStatus : 'open'
    const isMutationLocked = statusForLock === 'review' || statusForLock === 'closed';

    return {
        currentPeriod,
        periodStatus,
        setPeriodStatus,
        isBottomSheetOpen,
        openBottomSheet,
        closeBottomSheet,
        isMutationLocked,
        consensus,
        consensusLoading,
        loadConsensus,
        loadPeriod,
        isLoadingPeriod,
        selectedPeriodId,
        setSelectedPeriodId,
    };
}
