import { createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { APIError, authAPI, authorized } from './api'
import { clearSession, loadSession, refreshDelay, saveSession, toSession, type Session } from './session'

type Status = 'loading' | 'anonymous' | 'authenticated'

interface AuthContextValue {
  status: Status
  session: Session | null
  login(email: string, password: string): Promise<void>
  logout(): Promise<void>
  request<T>(path: string, init?: RequestInit): Promise<T>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(() => loadSession(window.sessionStorage))
  const [status, setStatus] = useState<Status>(() => session ? 'authenticated' : 'anonymous')
  const refreshInFlight = useRef<Promise<Session> | null>(null)

  const commit = useCallback((next: Session | null) => {
    setSession(next)
    setStatus(next ? 'authenticated' : 'anonymous')
    if (next) saveSession(window.sessionStorage, next)
    else clearSession(window.sessionStorage)
  }, [])

  const refresh = useCallback(async (current: Session): Promise<Session> => {
    if (!refreshInFlight.current) {
      refreshInFlight.current = authAPI.refresh(current.refreshToken)
        .then(toSession)
        .then(next => { commit(next); return next })
        .catch(error => { commit(null); throw error })
        .finally(() => { refreshInFlight.current = null })
    }
    return refreshInFlight.current
  }, [commit])

  useEffect(() => {
    if (!session) return
    const delay = refreshDelay(session)
    const timer = window.setTimeout(() => { void refresh(session).catch(() => undefined) }, delay)
    return () => window.clearTimeout(timer)
  }, [refresh, session])

  const login = useCallback(async (email: string, password: string) => {
    setStatus('loading')
    try { commit(toSession(await authAPI.login(email.trim().toLowerCase(), password))) }
    catch (error) { commit(null); throw error }
  }, [commit])

  const logout = useCallback(async () => {
    const current = session
    commit(null)
    if (current) await authAPI.logout(current.refreshToken).catch(() => undefined)
  }, [commit, session])

  const request = useCallback(async <T,>(path: string, init?: RequestInit): Promise<T> => {
    if (!session) throw new APIError(401, 'authentication_required', 'Authentication is required')
    try { return await authorized<T>(path, session.accessToken, init) }
    catch (error) {
      if (!(error instanceof APIError) || error.status !== 401) throw error
      const renewed = await refresh(session)
      return authorized<T>(path, renewed.accessToken, init)
    }
  }, [refresh, session])

  const value = useMemo(() => ({ status, session, login, logout, request }), [status, session, login, logout, request])
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const value = useContext(AuthContext)
  if (!value) throw new Error('useAuth must be used inside AuthProvider')
  return value
}
