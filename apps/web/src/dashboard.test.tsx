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
  history: [{ id: 1, snapshot_id: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', fingerprint: 'fingerprint', lifecycle_status: 'registration_open', data: { title: 'Pengadaan Integrasi Data', agency: 'Instansi Contoh', procurement_type: 'Jasa Lainnya', lifecycle_status: 'registration_open', budget_minor: 75000000000, currency: 'IDR' }, changed_fields: [], observed_at: '2026-08-09T02:00:00Z' }],
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
        : path.includes('/api/v1/team')
        ? { items: [{ user_id: 'user-1', email: 'admin@example.test', name: 'Administrator', roles: ['administrator'] }] }
        : path.includes('/api/v1/analytics/pipeline')
        ? { total: 173, active: 2, shortlisted: 1, assigned: 1, pursue: 1, pass: 0, undecided: 172, overdue_actions: 0, actions_due_seven_days: 1 }
        : path.includes('/api/v1/alert-rules') || path.includes('/api/v1/alerts')
        ? { items: [] }
        : path.includes('/api/v1/notification-preferences')
        ? { quiet_hours_enabled: false, quiet_hours_start: '22:00:00', quiet_hours_end: '07:00:00', timezone: 'Asia/Jakarta' }
        : path.includes('/api/v1/operations/dead-letters')
        ? { items: [] }
        : path.includes('/assignment')
        ? { assignment: null }
        : path.includes('/comments')
        ? { items: [{ id: 'comment-1', opportunity_id: opportunity.id, author_user_id: 'user-1', author_name: 'Administrator', body: 'Perlu validasi teknis.', created_at: '2026-08-10T00:00:00Z' }] }
        : path.includes('/shortlist')
        ? { opportunity_id: opportunity.id, shortlisted: false }
        : path.includes('/decision')
        ? { decision: null }
        : path.includes('/next-action')
        ? { next_action: null }
        : path.includes('/readiness')
        ? { opportunity_id: opportunity.id, version: 'readiness-v1', completed: 0, total: 1, items: [{ code: 'scope_validated', label: 'Scope and deliverables validated', completed: false, note: '' }] }
        : path.includes('/override')
        ? { override: null }
        : path.includes('/activity')
        ? { items: [{ id: 1, action: 'opportunity.shortlisted', actor_name: 'Administrator', metadata: {}, occurred_at: '2026-08-10T00:00:00Z' }] }
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
    expect(screen.getByRole('dialog')).toHaveTextContent('Initial observation')
    expect(screen.getByRole('dialog')).toHaveTextContent('Unassigned')
    expect(screen.getByRole('dialog')).toHaveTextContent('Perlu validasi teknis.')
    expect(screen.getByRole('button', { name: '☆ Add to shortlist' })).toBeInTheDocument()
    expect(screen.getByRole('dialog')).toHaveTextContent('Undecided')
    expect(screen.getByLabelText('Next action')).toBeInTheDocument()
    expect(screen.getByLabelText('Scope and deliverables validated')).toBeInTheDocument()
    expect(screen.getByLabelText('Override reason')).toBeInTheDocument()
    expect(screen.getByRole('dialog')).toHaveTextContent('opportunity · shortlisted')
    expect(screen.getByRole('link', { name: /open official source/i })).toHaveAttribute('href', opportunity.canonical_url)
    expect(await screen.findByText('Source health')).toBeInTheDocument()
    expect(screen.getByText('100/100')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Pipeline analytics' })).toBeInTheDocument()
    expect(screen.getByText('2 active of 173 total opportunities')).toBeInTheDocument()
  })

  it('switches to the needs-review queue endpoint', async () => {
    saveSession(window.sessionStorage, {
      accessToken: 'access', refreshToken: 'refresh', expiresAt: new Date(Date.now() + 600_000).toISOString(),
      user: { user_id: 'user-1', email: 'admin@example.test', name: 'Administrator', roles: ['administrator'] },
    })
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockImplementation(async input => new Response(JSON.stringify(String(input).includes('/api/v1/analytics/pipeline') ? { total: 0, active: 0, shortlisted: 0, assigned: 0, pursue: 0, pass: 0, undecided: 0, overdue_actions: 0, actions_due_seven_days: 0 } : String(input).includes('/api/v1/notification-preferences') ? { quiet_hours_enabled: false, quiet_hours_start: '22:00:00', quiet_hours_end: '07:00:00', timezone: 'Asia/Jakarta' } : String(input).includes('/api/v1/sources') || String(input).includes('/api/v1/saved-views') || String(input).includes('/api/v1/team') || String(input).includes('/api/v1/alert-rules') || String(input).includes('/api/v1/alerts') || String(input).includes('/api/v1/operations/dead-letters') ? { items: [] } : { items: [], page: 1, page_size: 20, total: 0, total_pages: 0 }), { status: 200, headers: { 'Content-Type': 'application/json' } }))
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
        : path.includes('/api/v1/analytics/pipeline') ? { total: 0, active: 0, shortlisted: 0, assigned: 0, pursue: 0, pass: 0, undecided: 0, overdue_actions: 0, actions_due_seven_days: 0 }
        : path.includes('/api/v1/notification-preferences') ? { quiet_hours_enabled: false, quiet_hours_start: '22:00:00', quiet_hours_end: '07:00:00', timezone: 'Asia/Jakarta' }
        : path.includes('/api/v1/sources') || path.includes('/api/v1/team') || path.includes('/api/v1/alert-rules') || path.includes('/api/v1/alerts') || path.includes('/api/v1/operations/dead-letters') ? { items: [] }
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

  it('completes the decision workflow and refreshes audit activity', async () => {
    saveSession(window.sessionStorage, {
      accessToken: 'access', refreshToken: 'refresh', expiresAt: new Date(Date.now() + 600_000).toISOString(),
      user: { user_id: 'user-1', email: 'operator@example.test', name: 'Operator', roles: ['operator'] },
    })
    const actions: string[] = []
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockImplementation(async (input, init) => {
      const path = String(input)
      const method = init?.method ?? 'GET'
      let body: unknown
      if (path.includes('/api/v1/sources') || path.includes('/api/v1/saved-views') || path.includes('/api/v1/team') || path.includes('/api/v1/alert-rules') || path.includes('/api/v1/alerts') || path.includes('/api/v1/operations/dead-letters')) body = { items: [] }
      else if (path.includes('/api/v1/analytics/pipeline')) body = { total: 1, active: 1, shortlisted: 0, assigned: 0, pursue: 0, pass: 0, undecided: 1, overdue_actions: 0, actions_due_seven_days: 0 }
      else if (path.includes('/api/v1/notification-preferences')) body = { quiet_hours_enabled: false, quiet_hours_start: '22:00:00', quiet_hours_end: '07:00:00', timezone: 'Asia/Jakarta' }
      else if (path.includes('/assignment')) body = { assignment: null }
      else if (path.includes('/comments')) {
        if (method === 'POST') actions.push('opportunity.comment_added')
        body = method === 'POST' ? { id: 'comment-1', opportunity_id: opportunity.id, author_user_id: 'user-1', author_name: 'Operator', body: 'Validate delivery capacity', created_at: '2026-08-13T00:00:00Z' } : { items: [] }
      }
      else if (path.includes('/readiness')) {
        if (method === 'PATCH') actions.push('opportunity.readiness_updated')
        body = { opportunity_id: opportunity.id, version: 'readiness-v1', completed: method === 'PATCH' ? 1 : 0, total: 1, items: [{ code: 'scope_validated', label: 'Scope and deliverables validated', completed: method === 'PATCH', note: '' }] }
      }
      else if (path.includes('/override')) {
        if (method === 'PATCH') actions.push('opportunity.override_updated')
        body = method === 'PATCH' ? { override: { opportunity_id: opportunity.id, recommendation: 'relevant', score_bp: null, reason: 'Reviewed by operator', overridden_by: 'user-1', overrider_name: 'Operator', created_at: '2026-08-13T00:00:00Z', updated_at: '2026-08-13T00:00:00Z' } } : { override: null }
      }
      else if (path.includes('/activity')) body = { items: actions.map((action, id) => ({ id: id + 1, action, actor_name: 'Operator', metadata: {}, occurred_at: '2026-08-13T00:00:00Z' })) }
      else if (path.includes('/shortlist')) {
        if (method === 'PATCH') actions.push('opportunity.shortlisted')
        body = { opportunity_id: opportunity.id, shortlisted: method === 'PATCH' }
      } else if (path.includes('/decision')) {
        if (method === 'PATCH') actions.push('opportunity.decision_changed')
        body = method === 'PATCH' ? { decision: { opportunity_id: opportunity.id, decision: 'pursue', reason: 'Strong capability fit', decided_by: 'user-1', decider_name: 'Operator', decided_at: '2026-08-13T00:00:00Z', updated_at: '2026-08-13T00:00:00Z' } } : { decision: null }
      } else if (path.includes('/next-action')) {
        if (method === 'PATCH') actions.push('opportunity.next_action_updated')
        body = method === 'PATCH' ? { next_action: { opportunity_id: opportunity.id, action: 'Validate technical scope', due_at: '2026-08-20T03:00:00Z', updated_by: 'user-1', updater_name: 'Operator', created_at: '2026-08-13T00:00:00Z', updated_at: '2026-08-13T00:00:00Z' } } : { next_action: null }
      } else if (path.includes(opportunity.id)) body = opportunity
      else body = { items: [opportunity], page: 1, page_size: 20, total: 1, total_pages: 1 }
      return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } })
    })

    const user = userEvent.setup()
    render(<AuthProvider><Dashboard /></AuthProvider>)
    await user.click(await screen.findByText(opportunity.title))
    await screen.findByRole('dialog')

    await user.click(screen.getByRole('button', { name: '☆ Add to shortlist' }))
    expect(await screen.findByRole('button', { name: '★ Shortlisted' })).toBeInTheDocument()

    await user.selectOptions(screen.getByLabelText('Pursue or pass decision'), 'pursue')
    await user.type(screen.getByLabelText('Decision reason'), 'Strong capability fit')
    await user.click(screen.getByRole('button', { name: 'Save decision' }))
    expect(await screen.findByRole('heading', { name: 'pursue' })).toBeInTheDocument()

    await user.type(screen.getByLabelText('Next action'), 'Validate technical scope')
    await user.type(screen.getByLabelText('Next action deadline'), '2026-08-20T10:00')
    await user.click(screen.getByRole('button', { name: 'Save action' }))
    expect(await screen.findByText('Validate technical scope')).toBeInTheDocument()

    await user.click(screen.getByLabelText('Scope and deliverables validated'))
    await user.selectOptions(screen.getByLabelText('Recommendation override'), 'relevant')
    await user.type(screen.getByLabelText('Override reason'), 'Reviewed by operator')
    await user.click(screen.getByRole('button', { name: 'Save override' }))
    await user.type(screen.getByLabelText('Internal comment'), 'Validate delivery capacity')
    await user.click(screen.getByRole('button', { name: 'Add comment' }))
    expect(await screen.findByText('Validate delivery capacity')).toBeInTheDocument()
    await vi.waitFor(() => expect(screen.getByRole('dialog')).toHaveTextContent('opportunity · comment added'))

    const patches = fetchMock.mock.calls.filter(([, init]) => init?.method === 'PATCH')
    expect(patches).toHaveLength(5)
    expect(JSON.parse(String(patches[0][1]?.body))).toEqual({ shortlisted: true })
    expect(JSON.parse(String(patches[1][1]?.body))).toEqual({ decision: 'pursue', reason: 'Strong capability fit' })
    expect(JSON.parse(String(patches[2][1]?.body))).toEqual({ action: 'Validate technical scope', due_at: '2026-08-20T03:00:00.000Z' })
    expect(JSON.parse(String(patches[3][1]?.body))).toEqual({ code: 'scope_validated', completed: true, note: '' })
    expect(JSON.parse(String(patches[4][1]?.body))).toEqual({ recommendation: 'relevant', score_bp: null, reason: 'Reviewed by operator', clear: false })
  })
})
