import { type FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { useAuth } from './auth'
import logoWhite from './assets/cankonix-logo-white.png'

export interface Opportunity {
  id: string
  package_identity: string
  title: string
  agency: string
  procurement_type: string
  lifecycle_status: string
  budget_minor?: number
  currency: string
  registration_deadline?: string
  canonical_url: string
  source_timestamp?: string
  first_seen_at: string
  last_seen_at: string
  source_id: string
  source_name: string
  version_count: number
  history?: Array<{ id: number; snapshot_id: string; fingerprint: string; lifecycle_status: string; data: { title: string; agency: string; procurement_type: string; lifecycle_status: string; budget_minor?: number; currency: string; registration_deadline?: string }; changed_fields: string[]; observed_at: string }>
  classification?: {
    disposition: 'relevant' | 'excluded' | 'needs_review'
    category_code?: string
    confidence_bp: number
    reason_code: string
    explanation: string
    auto_rejected: boolean
    evidence: Array<{ rule_code: string; kind: string; category_code?: string; keyword: string; source_field: string; excerpt: string }>
  }
  score?: {
    version: string
    total_score_bp: number
    confidence_bp: number
    needs_review: boolean
    reason_code: string
    explanation: string
    dimensions: Array<{ code: string; weight_bp: number; raw_score_bp: number; contribution: number; reason_code: string; explanation: string; evidence: string }>
  }
}

interface OpportunityPage { items: Opportunity[]; page: number; page_size: number; total: number; total_pages: number }

interface SourceHealth {
  id: string; name: string; status: string; health_score: number; policy_version: string
  last_success_at?: string; latest_run_status?: string; latest_run_error?: string
  latest_run_metrics: Record<string, number>; lag_seconds?: number; pending_runs: number
}

interface SavedView {
  id: string; name: string; created_at: string; updated_at: string
  filters: { q: string; status: string; source_id?: string; relevance: string; min_score_bp?: number; review_only: boolean }
}

const money = new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 })
const date = new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Jakarta' })

export function Dashboard() {
  const auth = useAuth()
  const [result, setResult] = useState<OpportunityPage>({ items: [], page: 1, page_size: 20, total: 0, total_pages: 0 })
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('active')
  const [sourceID, setSourceID] = useState('')
  const [relevance, setRelevance] = useState('all')
  const [minimumScore, setMinimumScore] = useState('')
  const [reviewOnly, setReviewOnly] = useState(false)
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState<Opportunity | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [sources, setSources] = useState<SourceHealth[]>([])
  const [sourceError, setSourceError] = useState('')
  const [savedViews, setSavedViews] = useState<SavedView[]>([])
  const [viewName, setViewName] = useState('')
  const [viewError, setViewError] = useState('')

  const load = useCallback(() => {
    setLoading(true)
    setError('')
    const params = new URLSearchParams({ q: search, status, relevance, page: String(page), page_size: '20' })
    if (sourceID) params.set('source_id', sourceID)
    if (minimumScore) params.set('min_score', String(Math.round(Number(minimumScore) * 100)))
    const path = reviewOnly ? '/api/v1/opportunities/review' : '/api/v1/opportunities'
    auth.request<OpportunityPage>(`${path}?${params}`)
      .then(setResult)
      .catch(() => setError('Feed peluang belum dapat dimuat. Coba beberapa saat lagi.'))
      .finally(() => setLoading(false))
  }, [auth.request, minimumScore, page, relevance, reviewOnly, search, sourceID, status])

  useEffect(load, [load])

  const loadSources = useCallback(() => {
    setSourceError('')
    auth.request<{ items: SourceHealth[] }>('/api/v1/sources')
      .then(body => setSources(body.items))
      .catch(() => setSourceError('Source health belum dapat dimuat.'))
  }, [auth.request])

  useEffect(loadSources, [loadSources])

  const loadSavedViews = useCallback(() => {
    auth.request<{ items: SavedView[] }>('/api/v1/saved-views')
      .then(body => setSavedViews(body.items))
      .catch(() => setViewError('Saved views belum dapat dimuat.'))
  }, [auth.request])

  useEffect(loadSavedViews, [loadSavedViews])

  function applySavedView(id: string) {
    const view = savedViews.find(item => item.id === id)
    if (!view) return
    setSearchInput(view.filters.q); setSearch(view.filters.q); setStatus(view.filters.status)
    setSourceID(view.filters.source_id ?? '')
    setRelevance(view.filters.relevance); setMinimumScore(view.filters.min_score_bp == null ? '' : String(view.filters.min_score_bp / 100))
    setReviewOnly(view.filters.review_only); setPage(1); setViewError('')
  }

  async function saveCurrentView(event: FormEvent) {
    event.preventDefault()
    const name = viewName.trim()
    if (!name) { setViewError('Nama saved view wajib diisi.'); return }
    try {
      const item = await auth.request<SavedView>('/api/v1/saved-views', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name, filters: { q: search, status, source_id: sourceID || undefined, relevance, min_score_bp: minimumScore ? Math.round(Number(minimumScore) * 100) : undefined, review_only: reviewOnly } }) })
      setSavedViews(items => [...items.filter(view => view.id !== item.id), item].sort((a, b) => a.name.localeCompare(b.name)))
      setViewName(''); setViewError('')
    } catch { setViewError('Saved view gagal disimpan. Gunakan nama yang unik.') }
  }

  async function deleteSavedView(id: string) {
    try {
      await auth.request<void>(`/api/v1/saved-views/${id}`, { method: 'DELETE' })
      setSavedViews(items => items.filter(item => item.id !== id)); setViewError('')
    } catch { setViewError('Saved view gagal dihapus.') }
  }

  const deadlineSoon = useMemo(() => result.items.filter(item => {
    if (!item.registration_deadline) return false
    const remaining = new Date(item.registration_deadline).getTime() - Date.now()
    return remaining > 0 && remaining <= 7 * 86400000
  }).length, [result.items])
  const needsReview = useMemo(() => result.items.filter(item => item.score?.needs_review).length, [result.items])

  function submitSearch(event: FormEvent) {
    event.preventDefault()
    setPage(1)
    setSearch(searchInput.trim())
  }

  async function openDetail(id: string) {
    setError('')
    try { setSelected(await auth.request<Opportunity>(`/api/v1/opportunities/${id}`)) }
    catch { setError('Detail peluang belum dapat dimuat.') }
  }

  return (
    <div className="shell">
      <aside>
        <div className="brand"><img className="brand-logo" src={logoWhite} alt="CANKONIX" /></div>
        <nav aria-label="Navigasi utama"><a href="#overview">Overview</a><a className="active" href="#opportunities">Opportunities</a><a href="#sources">Sources</a><a href="#operations">Operations</a></nav>
        <div className="profile"><strong>{auth.session?.user.name}</strong><span>{auth.session?.user.roles.join(' · ')}</span><button className="text-button" onClick={() => void auth.logout()}>Sign out</button></div>
        <p className="build">Opportunity Intelligence<br />Canonical feed · v0.3.0</p>
      </aside>
      <main>
        <header><div><p className="eyebrow">Decision support</p><h1>Opportunity feed</h1></div><div className="status"><i /> Collector connected</div></header>
        <section className="metrics" aria-label="Ringkasan">
          <article><p>Matched opportunities</p><strong>{result.total}</strong><small>Current filters</small></article>
          <article><p>Deadline &lt; 7 hari</p><strong>{deadlineSoon}</strong><small>Asia/Jakarta</small></article>
          <article><p>Visible packages</p><strong>{result.items.length}</strong><small>Page {result.page}</small></article>
          <article><p>Needs review</p><strong>{needsReview}</strong><small>Visible on this page</small></article>
        </section>

        <section className="feed" id="opportunities">
          <div className="saved-views">
            <label>Saved view<select aria-label="Apply saved view" defaultValue="" onChange={event => { applySavedView(event.target.value); event.target.value = '' }}><option value="" disabled>Choose view…</option>{savedViews.map(view => <option key={view.id} value={view.id}>{view.name}</option>)}</select></label>
            <form onSubmit={saveCurrentView}><input aria-label="Saved view name" maxLength={80} value={viewName} onChange={event => setViewName(event.target.value)} placeholder="Name current filters…" /><button type="submit">Save view</button></form>
            {savedViews.length > 0 && <div className="saved-view-chips">{savedViews.map(view => <span key={view.id}>{view.name}<button aria-label={`Delete saved view ${view.name}`} onClick={() => void deleteSavedView(view.id)}>×</button></span>)}</div>}
            {viewError && <small role="alert">{viewError}</small>}
          </div>
          <div className="feed-toolbar">
            <form onSubmit={submitSearch} role="search"><input aria-label="Cari peluang" value={searchInput} onChange={event => setSearchInput(event.target.value)} placeholder="Cari paket, instansi, atau kode…" /><button type="submit">Search</button></form>
            <label>Status<select aria-label="Filter status" value={status} onChange={event => { setStatus(event.target.value); setPage(1) }}><option value="active">Active registration</option><option value="all">All lifecycle</option><option value="announced">Announced</option><option value="registration_open">Registration open</option><option value="registration_closed">Registration closed</option><option value="evaluation">Evaluation</option><option value="awarded">Awarded</option><option value="contracted">Contracted</option><option value="completed">Completed</option><option value="cancelled">Cancelled</option></select></label>
            <label>Source<select aria-label="Filter source" value={sourceID} onChange={event => { setSourceID(event.target.value); setPage(1) }}><option value="">All sources</option>{sources.map(source => <option key={source.id} value={source.id}>{source.name}</option>)}</select></label>
            <label>Relevance<select aria-label="Filter relevance" value={relevance} onChange={event => { setRelevance(event.target.value); setPage(1) }}><option value="all">All</option><option value="relevant">Relevant</option><option value="needs_review">Ambiguous</option><option value="excluded">Excluded</option></select></label>
            <label>Min score<input className="score-input" aria-label="Minimum score" type="number" min="0" max="100" value={minimumScore} onChange={event => { setMinimumScore(event.target.value); setPage(1) }} placeholder="0–100" /></label>
            <label className="review-toggle"><input aria-label="Review queue only" type="checkbox" checked={reviewOnly} onChange={event => { setReviewOnly(event.target.checked); setPage(1) }} />Review queue</label>
          </div>
          {error && <div className="feed-error" role="alert">{error}<button onClick={load}>Retry</button></div>}
          {loading ? <div className="feed-state">Loading canonical opportunities…</div> : result.items.length === 0 ? <div className="feed-state"><strong>No opportunities found.</strong><span>Adjust the search or lifecycle filter.</span></div> :
            <div className="opportunity-list">{result.items.map(item => <button className="opportunity-row" key={item.id} onClick={() => void openDetail(item.id)}>
              <span className="opportunity-code">{item.package_identity}</span><span className="opportunity-main"><strong>{item.title}</strong><small>{item.agency} · {item.procurement_type}</small></span>
              <span className="opportunity-budget">{item.budget_minor == null ? 'Nilai belum tersedia' : money.format(item.budget_minor)}<small>{item.registration_deadline ? `Tutup ${date.format(new Date(item.registration_deadline))}` : 'Deadline belum tersedia'}</small></span>
              <span className={`score-badge ${item.score?.needs_review ? 'review' : ''}`}>{item.score ? `${(item.score.total_score_bp / 100).toFixed(0)} score` : 'Unscored'}<small>{item.classification?.disposition.replaceAll('_', ' ') ?? 'unclassified'}</small></span>
              <span className={`lifecycle ${item.lifecycle_status}`}>{item.lifecycle_status.replaceAll('_', ' ')}</span>
            </button>)}</div>}
          <div className="pagination"><button disabled={page <= 1} onClick={() => setPage(value => value - 1)}>Previous</button><span>Page {result.page} of {Math.max(result.total_pages, 1)}</span><button disabled={page >= result.total_pages} onClick={() => setPage(value => value + 1)}>Next</button></div>
        </section>
        <section className="operations" id="operations">
          <div className="section-heading"><div><p className="eyebrow">Operations</p><h2>Source health</h2></div><button onClick={loadSources}>Refresh</button></div>
          {sourceError ? <div className="feed-error" role="alert">{sourceError}</div> :
            <div className="source-grid">{sources.map(source => <article key={source.id}>
              <div className="source-title"><strong>{source.name}</strong><span className={`health health-${source.status}`}>{source.status}</span></div>
              <dl><div><dt>Health</dt><dd>{source.health_score}/100</dd></div><div><dt>Latest run</dt><dd>{source.latest_run_status || 'never'}</dd></div><div><dt>Pending</dt><dd>{source.pending_runs}</dd></div><div><dt>Lag</dt><dd>{source.lag_seconds == null ? 'never' : `${Math.floor(source.lag_seconds / 60)} min`}</dd></div></dl>
              <small>{source.last_success_at ? `Last success ${date.format(new Date(source.last_success_at))}` : 'No successful run yet'}</small>
              {source.latest_run_error && <p className="source-error">{source.latest_run_error}</p>}
              <p className="source-metrics">{source.latest_run_metrics?.unique_candidates ?? source.latest_run_metrics?.candidates ?? 0} candidates · {source.latest_run_metrics?.new_snapshots ?? 0} new</p>
            </article>)}</div>}
        </section>
      </main>
      {selected && <div className="detail-backdrop" role="presentation" onMouseDown={() => setSelected(null)}><section className="detail-panel" role="dialog" aria-modal="true" aria-labelledby="detail-title" onMouseDown={event => event.stopPropagation()}>
        <button className="detail-close" aria-label="Tutup detail" onClick={() => setSelected(null)}>×</button><p className="eyebrow">{selected.package_identity}</p><h2 id="detail-title">{selected.title}</h2><span className={`lifecycle ${selected.lifecycle_status}`}>{selected.lifecycle_status.replaceAll('_', ' ')}</span>
        <dl><div><dt>Instansi</dt><dd>{selected.agency}</dd></div><div><dt>Jenis pengadaan</dt><dd>{selected.procurement_type}</dd></div><div><dt>Nilai paket</dt><dd>{selected.budget_minor == null ? 'Belum tersedia' : money.format(selected.budget_minor)}</dd></div><div><dt>Batas pendaftaran</dt><dd>{selected.registration_deadline ? date.format(new Date(selected.registration_deadline)) : 'Belum tersedia'}</dd></div><div><dt>Sumber</dt><dd>{selected.source_name}</dd></div><div><dt>Riwayat terdeteksi</dt><dd>{selected.version_count} versi · pertama {date.format(new Date(selected.first_seen_at))}</dd></div></dl>
        {selected.classification && <section className="analysis-card"><div className="analysis-heading"><div><p className="eyebrow">Deterministic classification</p><h3>{selected.classification.disposition.replaceAll('_', ' ')}</h3></div><strong>{(selected.classification.confidence_bp / 100).toFixed(0)}%</strong></div><p>{selected.classification.explanation}</p>{selected.classification.auto_rejected && <div className="exclusion-warning">Approved deterministic exclusion · not an AI recommendation</div>}<ul>{selected.classification.evidence.map((evidence, index) => <li key={`${evidence.rule_code}-${index}`}><strong>{evidence.rule_code}</strong><span>“{evidence.keyword}” in {evidence.source_field}</span></li>)}</ul></section>}
        {selected.score && <section className="analysis-card"><div className="analysis-heading"><div><p className="eyebrow">Explainable score</p><h3>{(selected.score.total_score_bp / 100).toFixed(0)} / 100</h3></div><strong>{selected.score.needs_review ? 'Review' : 'Ready'}</strong></div><p>{selected.score.explanation}</p><div className="score-breakdown">{selected.score.dimensions.map(dimension => <div key={dimension.code}><span>{dimension.code.replaceAll('_', ' ')}</span><progress max="10000" value={dimension.raw_score_bp} /><strong>{(dimension.raw_score_bp / 100).toFixed(0)}</strong><small>{dimension.explanation}</small></div>)}</div></section>}
        <section className="timeline"><p className="eyebrow">Change timeline</p><h3>{selected.history?.length ?? 0} observations</h3>{selected.history?.map((version, index) => <article key={version.id}><i /><div><strong>{index === selected.history!.length - 1 ? 'Initial observation' : version.changed_fields.length ? version.changed_fields.map(field => field.replaceAll('_', ' ')).join(' · ') : 'Observed again'}</strong><time>{date.format(new Date(version.observed_at))}</time><span>{version.lifecycle_status.replaceAll('_', ' ')} · snapshot {version.snapshot_id.slice(0, 8)}</span></div></article>)}</section>
        <section className="ai-separation"><strong>AI recommendation</strong><span>Belum dijalankan. Hasil di atas berasal dari deterministic rules dan scoring.</span></section>
        <a className="official-link" href={selected.canonical_url} target="_blank" rel="noreferrer">Open official source ↗</a><p className="detail-note">Verify all procurement details on the official source before making a decision.</p>
      </section></div>}
    </div>
  )
}
