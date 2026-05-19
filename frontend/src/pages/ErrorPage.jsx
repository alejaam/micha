import { useRouteError, Link } from 'react-router-dom'

/**
 * ErrorPage — React Router v6 errorElement fallback.
 * Displays a user-friendly error message when routes throw errors.
 */
export function ErrorPage() {
    const error = useRouteError()
    
    console.error('Route error:', error)

    const errorMessage = error?.message || error?.statusText || 'Algo salió mal'
    const isNotFound = error?.status === 404

    return (
        <div className="page" style={{ 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            minHeight: '100vh',
            padding: '1rem'
        }}>
            <section className="card" style={{ maxWidth: 480, textAlign: 'center' }}>
                <p style={{ fontSize: '3rem', marginBottom: '1rem' }} aria-hidden>
                    {isNotFound ? '🔍' : '💥'}
                </p>
                <h1 className="sectionTitle">
                    {isNotFound ? 'Página no encontrada' : 'Algo salió mal'}
                </h1>
                <p className="text-secondary" style={{ marginBottom: '1.5rem' }}>
                    {isNotFound 
                        ? 'La página que buscas no existe.' 
                        : `Error: ${errorMessage}`}
                </p>
                <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'center' }}>
                    <Link to="/" className="btn btnPrimary">
                        Ir al inicio
                    </Link>
                    <button 
                        type="button" 
                        className="btn btnGhost"
                        onClick={() => window.location.reload()}
                    >
                        Recargar
                    </button>
                </div>
            </section>
        </div>
    )
}
