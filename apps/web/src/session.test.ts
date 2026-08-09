import { describe, expect, it } from 'vitest'
import { clearSession, loadSession, refreshDelay, saveSession, type Session } from './session'

const session: Session = { accessToken: 'access', refreshToken: 'refresh', expiresAt: '2026-08-09T12:15:00Z', user: { user_id: 'user-1', email: 'admin@example.test', name: 'Admin', roles: ['administrator'] } }

function storage(): Storage {
  const values = new Map<string, string>()
  return { length: 0, clear: () => values.clear(), getItem: key => values.get(key) ?? null, key: () => null, removeItem: key => { values.delete(key) }, setItem: (key, value) => { values.set(key, value) } }
}

describe('session persistence', () => {
  it('round trips and clears a valid session', () => { const s = storage(); saveSession(s, session); expect(loadSession(s)).toEqual(session); clearSession(s); expect(loadSession(s)).toBeNull() })
  it('rejects malformed persisted data', () => { const s = storage(); s.setItem('cankonix.oi.session.v1', '{bad'); expect(loadSession(s)).toBeNull() })
  it('refreshes one minute before expiration', () => { expect(refreshDelay(session, Date.parse('2026-08-09T12:10:00Z'))).toBe(240_000) })
})
