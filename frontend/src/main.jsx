import React from 'react'
import ReactDOM from 'react-dom/client'
import { AuthProvider } from './context/AuthContext'
import { ThemeProvider } from './context/ThemeContext'
import AppShell from './router'
import './styles.css'
import './styles/auth-onboarding.css'

ReactDOM.createRoot(document.getElementById('root')).render(
    <React.StrictMode>
        <ThemeProvider>
            <AuthProvider>
                <AppShell />
            </AuthProvider>
        </ThemeProvider>
    </React.StrictMode>,
)
