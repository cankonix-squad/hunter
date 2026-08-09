export interface Principal {
  user_id: string
  email: string
  name: string
  roles: string[]
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  token_type: 'Bearer'
  expires_at: string
  user: Principal
}

export interface Session {
  accessToken: string
  refreshToken: string
  expiresAt: string
  user: Principal
}

const key = 'cankonix.oi.session.v1'

export function toSession(tokens: TokenResponse): Session {
  return {
    accessToken: tokens.access_token,
    refreshToken: tokens.refresh_token,
    expiresAt: tokens.expires_at,
    user: tokens.user,
  }
}

export function loadSession(storage: Storage): Session | null {
  const raw = storage.getItem(key)
  if (!raw) return null
  try {
    const value = JSON.parse(raw) as Partial<Session>
    if (!value.accessToken || !value.refreshToken || !value.expiresAt || !value.user?.user_id) {
      storage.removeItem(key)
      return null
    }
    return value as Session
  } catch {
    storage.removeItem(key)
    return null
  }
}

export function saveSession(storage: Storage, session: Session): void {
  storage.setItem(key, JSON.stringify(session))
}

export function clearSession(storage: Storage): void {
  storage.removeItem(key)
}

export function refreshDelay(session: Session, now = Date.now()): number {
  const expires = Date.parse(session.expiresAt)
  if (!Number.isFinite(expires)) return 0
  return Math.max(0, expires - now - 60_000)
}
