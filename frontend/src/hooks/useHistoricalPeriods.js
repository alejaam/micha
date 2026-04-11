import { useState, useEffect } from 'react';

/**
 * Hook for managing historical periods data.
 * Falls back to mocked or derived data if backend endpoints are not ready.
 */
export function useHistoricalPeriods(householdId) {
    const [historicalData, setHistoricalData] = useState([]);
    const [isLoading, setIsLoading] = useState(false);

    useEffect(() => {
        if (!householdId) return;

        // TODO: Implement Phase 5 historical data fetching or mocking
        setIsLoading(true);
        setTimeout(() => {
            setHistoricalData([]);
            setIsLoading(false);
        }, 500);

    }, [householdId]);

    return {
        historicalData,
        isLoading
    };
}
