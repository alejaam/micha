import '@testing-library/jest-dom';

// jsdom does not implement window.matchMedia; components like ThemeProvider
// rely on it to resolve the system color scheme preference.
if (typeof window.matchMedia !== 'function') {
    window.matchMedia = (query) => ({
        matches: false,
        media: query,
        addEventListener: () => {},
        removeEventListener: () => {},
    });
}
