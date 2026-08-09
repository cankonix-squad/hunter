// @vitest-environment jsdom
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { AuthProvider } from './auth'
import { LoginPage } from './login'

const tokens = {
  access_token: 'access', refresh_token: 'refresh', token_type: 'Bearer', expires_at: new Date(Date.now() + 600_000).toISOString(),
  user: { user_id: 'user-1', email: 'admin@example.test', name: 'Administrator', roles: ['administrator'] },
}

afterEach(() => { cleanup(); vi.restoreAllMocks(); window.sessionStorage.clear() })

describe('LoginPage', () => {
  it('submits normalized credentials and stores the session', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(tokens), { status: 200, headers: { 'Content-Type': 'application/json' } }))
    render(<AuthProvider><LoginPage /></AuthProvider>)
    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Email'), 'ADMIN@EXAMPLE.TEST')
    await user.type(screen.getByLabelText('Password'), 'a-secure-password')
    await user.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/auth/login', expect.objectContaining({ body: JSON.stringify({ email: 'admin@example.test', password: 'a-secure-password' }) }))
    expect(window.sessionStorage.length).toBe(1)
  })

  it('shows a safe invalid-credential error', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ error: { code: 'invalid_credentials', message: 'invalid' } }), { status: 401, headers: { 'Content-Type': 'application/json' } }))
    render(<AuthProvider><LoginPage /></AuthProvider>)
    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Email'), 'admin@example.test')
    await user.type(screen.getByLabelText('Password'), 'wrong-password')
    await user.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Email atau password tidak cocok.')
  })
})
