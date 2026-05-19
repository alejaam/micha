import { Component } from 'react'

/**
 * ErrorBoundary — Catches JavaScript errors in child components and displays
 * a fallback UI instead of crashing the entire app.
 */
export class ErrorBoundary extends Component {
    constructor(props) {
        super(props)
        this.state = { hasError: false, error: null }
    }

    static getDerivedStateFromError(error) {
        return { hasError: true, error }
    }

    componentDidCatch(error, errorInfo) {
        console.error('ErrorBoundary caught error:', error, errorInfo)
    }

    handleReset = () => {
        this.setState({ hasError: false, error: null })
    }

    render() {
        if (this.state.hasError) {
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
                            🔧
                        </p>
                        <h1 className="sectionTitle">Error inesperado</h1>
                        <p className="text-secondary" style={{ marginBottom: '1.5rem' }}>
                            Algo salió mal en la aplicación. Intenta recargar la página.
                        </p>
                        {this.state.error?.message && (
                            <pre style={{ 
                                fontSize: '0.75rem', 
                                background: 'var(--surface-2)', 
                                padding: '0.75rem',
                                borderRadius: '0.5rem',
                                marginBottom: '1rem',
                                overflow: 'auto'
                            }}>
                                {this.state.error.message}
                            </pre>
                        )}
                        <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'center' }}>
                            <button 
                                type="button" 
                                className="btn btnPrimary"
                                onClick={this.handleReset}
                            >
                                Intentar de nuevo
                            </button>
                            <button 
                                type="button" 
                                className="btn btnGhost"
                                onClick={() => window.location.reload()}
                            >
                                Recargar página
                            </button>
                        </div>
                    </section>
                </div>
            )
        }

        return this.props.children
    }
}
