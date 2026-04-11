import { useState } from 'react';

/**
 * Hook for managing the dashboard UI contextual state (ribbon status, active views).
 */
export function useDashboardUxState(initialStatus = 'open') {
    // periodStatus can be: 'open', 'review', 'closed'
    const [periodStatus, setPeriodStatus] = useState(initialStatus);
    const [isBottomSheetOpen, setIsBottomSheetOpen] = useState(false);

    const openBottomSheet = () => setIsBottomSheetOpen(true);
    const closeBottomSheet = () => setIsBottomSheetOpen(false);

    // Business rule: lock mutations during 'review'
    const isMutationLocked = periodStatus === 'review' || periodStatus === 'closed';

    return {
        periodStatus,
        setPeriodStatus,
        isBottomSheetOpen,
        openBottomSheet,
        closeBottomSheet,
        isMutationLocked
    };
}
