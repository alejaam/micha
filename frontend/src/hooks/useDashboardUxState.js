import { useEffect, useState } from 'react';
import { getCurrentPeriod } from '../api';

/**
 * Hook for managing the dashboard UI contextual state (ribbon status, active views).
 */
const PERIOD_STATUS_MAP = {
    open: {
        stateLabel: '[OPEN]',
        description: 'Periodo abierto — puedes registrar y editar gastos.',
    },
    review: {
        stateLabel: '[REVIEW]',
        description: 'Periodo en revisión — las acciones de edición están bloqueadas temporalmente.',
    },
    closed: {
        stateLabel: '[CLOSED]',
        description: 'Periodo cerrado — ya no se permiten cambios en gastos.',
    },
}

export function buildRibbonState(status = 'open') {
    const normalizedStatus = PERIOD_STATUS_MAP[status] ? status : 'open'
    return {
        status: normalizedStatus,
        ...PERIOD_STATUS_MAP[normalizedStatus],
    }
}

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

    const loadPeriod = async () => {
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
    }

    useEffect(() => {
        loadPeriod();
    }, [householdId]);

    const openBottomSheet = () => setIsBottomSheetOpen(true);
    const closeBottomSheet = () => setIsBottomSheetOpen(false);

    // Business rule: lock mutations during 'review'
    const normalizedPeriodStatus = buildRibbonState(periodStatus).status
    const isMutationLocked = normalizedPeriodStatus === 'review' || normalizedPeriodStatus === 'closed';
    const consensus = buildConsensusState({ approved: 0, total: 0, source: 'mock' })

    return {
        currentPeriod,
        periodStatus,
        setPeriodStatus,
        isBottomSheetOpen,
        openBottomSheet,
        closeBottomSheet,
        isMutationLocked,
        consensus,
        loadPeriod,
        isLoadingPeriod,
    };
}
