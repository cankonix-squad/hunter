// @vitest-environment jsdom
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { AuthProvider } from './auth'
import { Dashboard, type Opportunity } from './dashboard'
import { saveSession } from './session'

const opportunity: Opportunity = {
  id: '8f75b02b-d04f-43bc-99bd-f67d76d9e1dd', package_identity: 'SPSE-001',
  title: 'Pengadaan Integrasi Data', agency: 'Instansi Contoh', procurement_type: 'Jasa Lainnya',
  lifecycle_status: 'registration_open', budget_minor: 75000000000, currency: 'IDR',
  registration_deadline: '2026-08-20T09:00:00Z', canonical_url: 'https://lpse.example.test/package/1',
  first_seen_at: '2026-08-09T02:00:00Z', last_seen_at: '2026-08-09T02:00:00Z',
  source_id: 'source-1', source_name: 'LPSE Contoh', version_count: 1,
}

afterEach(() => { cleanup(); vi.restoreAllMocks(); window.sessionStorage.clear() })

describe('Dashboard', () => {
  it('renders canonical feed and opens auditable detail', async () => {
    saveSession(window.sessionStorage, {
      accessToken: 'access', refreshToken: 'refresh', expiresAt: new Date(Date.now() + 600_000).toISOString(),
      user: { user_id: 'user-1', email: 'admin@example.test', name: 'Administrator', roles: ['administrator'] },
    })
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockImplementation(async input => {
      const path = String(input)
      const body = path.includes(opportunity.id)
        ? opportunity
        : { items: [opportunity], page: 1, page_size: 20, total: 1, total_pages: 1 }
      return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
    })
    render(<AuthProvider><Dashboard /></AuthProvider>)
    expect(await screen.findByText('Pengadaan Integrasi Data')).toBeInTheDocument()
    expect(fetchMock.mock.calls[0][0]).toContain('/api/v1/opportunities?')
    await userEvent.click(screen.getByText('Pengadaan Integrasi Data'))
    expect(await screen.findByRole('dialog')).toHaveTextContent('LPSE Contoh')
    expect(screen.getByRole('link', { name: /open official source/i })).toHaveAttribute('href', opportunity.canonical_url)
  })
})
