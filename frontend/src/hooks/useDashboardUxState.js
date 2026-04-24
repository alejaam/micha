import { useState } from 'react';

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

export function useDashboardUxState(initialStatus = 'open') {
    // periodStatus can be: 'open', 'review', 'closed'
    const [periodStatus, setPeriodStatus] = useState(buildRibbonState(initialStatus).status);
    const [isBottomSheetOpen, setIsBottomSheetOpen] = useState(false);

    const openBottomSheet = () => setIsBottomSheetOpen(true);
    const closeBottomSheet = () => setIsBottomSheetOpen(false);

    // Business rule: lock mutations during 'review'
    const normalizedPeriodStatus = buildRibbonState(periodStatus).status
    const isMutationLocked = normalizedPeriodStatus === 'review' || normalizedPeriodStatus === 'closed';
    const consensus = buildConsensusState({ approved: 0, total: 0, source: 'mock' })

    return {
        periodStatus,
        setPeriodStatus,
        isBottomSheetOpen,
        openBottomSheet,
        closeBottomSheet,
        isMutationLocked,
        consensus,
    };
}
