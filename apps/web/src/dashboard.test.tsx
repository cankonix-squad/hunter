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
  classification: { disposition: 'relevant', category_code: 'data-ai', confidence_bp: 9000,
    reason_code: 'include_data', explanation: 'Paket sesuai kapabilitas data.', auto_rejected: false,
    evidence: [{ rule_code: 'include_data', kind: 'include', category_code: 'data-ai', keyword: 'integrasi data', source_field: 'title', excerpt: 'Pengadaan Integrasi Data' }] },
  score: { version: 'explainable-v1', total_score_bp: 8000, confidence_bp: 9000, needs_review: false,
    reason_code: 'score_ready', explanation: 'Score deterministik lengkap.', dimensions: [
      { code: 'capability_fit', weight_bp: 4000, raw_score_bp: 9000, contribution: 3600, reason_code: 'deterministic_fit', explanation: 'Kapabilitas sesuai.', evidence: 'include_data' },
    ] },
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
      const body = path.includes('/api/v1/sources')
        ? { items: [{ id: 'source-1', name: 'LPSE Contoh', status: 'active', health_score: 100, policy_version: 'v1', latest_run_status: 'succeeded', latest_run_metrics: { candidates: 1, new_snapshots: 1 }, lag_seconds: 60, pending_runs: 0 }] }
        : path.includes('/api/v1/saved-views')
        ? { items: [] }
        : path.includes(opportunity.id)
        ? opportunity
        : { items: [opportunity], page: 1, page_size: 20, total: 1, total_pages: 1 }
      return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
    })
    render(<AuthProvider><Dashboard /></AuthProvider>)
    expect(await screen.findByText('Pengadaan Integrasi Data')).toBeInTheDocument()
    expect(fetchMock.mock.calls[0][0]).toContain('/api/v1/opportunities?')
    await userEvent.click(screen.getByText('Pengadaan Integrasi Data'))
    expect(await screen.findByRole('dialog')).toHaveTextContent('LPSE Contoh')
    expect(screen.getByRole('dialog')).toHaveTextContent('80 / 100')
    expect(screen.getByRole('dialog')).toHaveTextContent('include_data')
    expect(screen.getByRole('dialog')).toHaveTextContent('Belum dijalankan')
    expect(screen.getByRole('link', { name: /open official source/i })).toHaveAttribute('href', opportunity.canonical_url)
    expect(await screen.findByText('Source health')).toBeInTheDocument()
    expect(screen.getByText('100/100')).toBeInTheDocument()
  })

  it('switches to the needs-review queue endpoint', async () => {
    saveSession(window.sessionStorage, {
      accessToken: 'access', refreshToken: 'refresh', expiresAt: new Date(Date.now() + 600_000).toISOString(),
      user: { user_id: 'user-1', email: 'admin@example.test', name: 'Administrator', roles: ['administrator'] },
    })
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockImplementation(async input => new Response(JSON.stringify(String(input).includes('/api/v1/sources') || String(input).includes('/api/v1/saved-views') ? { items: [] } : { items: [], page: 1, page_size: 20, total: 0, total_pages: 0 }), { status: 200, headers: { 'Content-Type': 'application/json' } }))
    render(<AuthProvider><Dashboard /></AuthProvider>)
    await screen.findByText('No opportunities found.')
    await userEvent.click(screen.getByLabelText('Review queue only'))
    await vi.waitFor(() => expect(fetchMock.mock.calls.some(call => String(call[0]).includes('/api/v1/opportunities/review?'))).toBe(true))
  })

  it('applies all filters from a saved view', async () => {
    saveSession(window.sessionStorage, {
      accessToken: 'access', refreshToken: 'refresh', expiresAt: new Date(Date.now() + 600_000).toISOString(),
      user: { user_id: 'user-1', email: 'admin@example.test', name: 'Administrator', roles: ['administrator'] },
    })
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockImplementation(async input => {
      const path = String(input)
      const body = path.includes('/api/v1/saved-views')
        ? { items: [{ id: 'view-1', name: 'Relevant closed', filters: { q: 'data', status: 'completed', relevance: 'relevant', min_score_bp: 7500, review_only: true }, created_at: '2026-08-10T00:00:00Z', updated_at: '2026-08-10T00:00:00Z' }] }
        : path.includes('/api/v1/sources') ? { items: [] }
        : { items: [], page: 1, page_size: 20, total: 0, total_pages: 0 }
      return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
    })
    render(<AuthProvider><Dashboard /></AuthProvider>)
    await screen.findByRole('option', { name: 'Relevant closed' })
    await userEvent.selectOptions(screen.getByLabelText('Apply saved view'), 'view-1')
    await vi.waitFor(() => expect(fetchMock.mock.calls.some(call => {
      const path = String(call[0])
      return path.includes('/api/v1/opportunities/review?') && path.includes('q=data') && path.includes('status=completed') && path.includes('relevance=relevant') && path.includes('min_score=7500')
    })).toBe(true))
  })
})
