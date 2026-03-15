import { createContext, useCallback, useContext, useEffect, useState } from 'react'
import { getMe, loginUser, registerUser, setAuthToken } from '../api'

const AUTH_STORAGE_KEY = 'micha_token'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
    const [token, setToken] = useState(() => localStorage.getItem(AUTH_STORAGE_KEY) ?? '')
    const [user, setUser] = useState(null) // { user_id, email }
    const [loadingUser, setLoadingUser] = useState(false)

    const isAuthenticated = token.trim() !== ''

    // Sync token to api module on every change
    useEffect(() => {
        setAuthToken(token)
    }, [token])

    // Fetch /v1/auth/me whenever we have a token
    useEffect(() => {
        if (!token.trim()) {
            setUser(null)
            return
        }

        let active = true
        setLoadingUser(true)
        getMe()
            .then((data) => {
                if (active) setUser(data ?? null)
            })
            .catch(() => {
                if (active) setUser(null)
            })
            .finally(() => {
                if (active) setLoadingUser(false)
            })

        return () => {
            active = false
        }
    }, [token])

    const logout = useCallback((reason = '') => {
        localStorage.removeItem(AUTH_STORAGE_KEY)
        setToken('')
        setAuthToken('')
        setUser(null)
        return reason // caller can use this to show a message
    }, [])

    const login = useCallback(async ({ email, password }) => {
        const out = await loginUser({ email, password })
        const nextToken = out?.token ?? ''
        if (!nextToken) throw new Error('login succeeded but token was not returned')
        localStorage.setItem(AUTH_STORAGE_KEY, nextToken)
        setAuthToken(nextToken)
        setToken(nextToken)
    }, [])

    const register = useCallback(async ({ email, password }) => {
        await registerUser({ email, password })
    }, [])

    const handleProtectedError = useCallback(
        (err) => {
            if (err?.code === 'UNAUTHORIZED') {
                logout('Session expired. Sign in again.')
                return true
            }
            return false
        },
        [logout],
    )

    return (
        <AuthContext.Provider
            value={{
                token,
                user,
                loadingUser,
                isAuthenticated,
                login,
                register,
                logout,
                handleProtectedError,
            }}
        >
            {children}
        </AuthContext.Provider>
    )
}

export function useAuth() {
    const ctx = useContext(AuthContext)
    if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>')
    return ctx
}
