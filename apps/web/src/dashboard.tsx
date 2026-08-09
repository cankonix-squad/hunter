import { type FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
import { useAuth } from './auth'

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

const money = new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 })
const date = new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Jakarta' })

export function Dashboard() {
  const auth = useAuth()
  const [result, setResult] = useState<OpportunityPage>({ items: [], page: 1, page_size: 20, total: 0, total_pages: 0 })
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('active')
  const [relevance, setRelevance] = useState('all')
  const [minimumScore, setMinimumScore] = useState('')
  const [reviewOnly, setReviewOnly] = useState(false)
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState<Opportunity | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    setLoading(true)
    setError('')
    const params = new URLSearchParams({ q: search, status, relevance, page: String(page), page_size: '20' })
    if (minimumScore) params.set('min_score', String(Math.round(Number(minimumScore) * 100)))
    const path = reviewOnly ? '/api/v1/opportunities/review' : '/api/v1/opportunities'
    auth.request<OpportunityPage>(`${path}?${params}`)
      .then(setResult)
      .catch(() => setError('Feed peluang belum dapat dimuat. Coba beberapa saat lagi.'))
      .finally(() => setLoading(false))
  }, [auth.request, minimumScore, page, relevance, reviewOnly, search, status])

  useEffect(load, [load])

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
        <div className="brand"><span>CX</span><strong>CANKONIX</strong></div>
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
          <div className="feed-toolbar">
            <form onSubmit={submitSearch} role="search"><input aria-label="Cari peluang" value={searchInput} onChange={event => setSearchInput(event.target.value)} placeholder="Cari paket, instansi, atau kode…" /><button type="submit">Search</button></form>
            <label>Status<select aria-label="Filter status" value={status} onChange={event => { setStatus(event.target.value); setPage(1) }}><option value="active">Active registration</option><option value="all">All lifecycle</option><option value="announced">Announced</option><option value="registration_open">Registration open</option><option value="registration_closed">Registration closed</option><option value="evaluation">Evaluation</option><option value="awarded">Awarded</option><option value="completed">Completed</option><option value="cancelled">Cancelled</option></select></label>
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
      </main>
      {selected && <div className="detail-backdrop" role="presentation" onMouseDown={() => setSelected(null)}><section className="detail-panel" role="dialog" aria-modal="true" aria-labelledby="detail-title" onMouseDown={event => event.stopPropagation()}>
        <button className="detail-close" aria-label="Tutup detail" onClick={() => setSelected(null)}>×</button><p className="eyebrow">{selected.package_identity}</p><h2 id="detail-title">{selected.title}</h2><span className={`lifecycle ${selected.lifecycle_status}`}>{selected.lifecycle_status.replaceAll('_', ' ')}</span>
        <dl><div><dt>Instansi</dt><dd>{selected.agency}</dd></div><div><dt>Jenis pengadaan</dt><dd>{selected.procurement_type}</dd></div><div><dt>Nilai paket</dt><dd>{selected.budget_minor == null ? 'Belum tersedia' : money.format(selected.budget_minor)}</dd></div><div><dt>Batas pendaftaran</dt><dd>{selected.registration_deadline ? date.format(new Date(selected.registration_deadline)) : 'Belum tersedia'}</dd></div><div><dt>Sumber</dt><dd>{selected.source_name}</dd></div><div><dt>Riwayat terdeteksi</dt><dd>{selected.version_count} versi · pertama {date.format(new Date(selected.first_seen_at))}</dd></div></dl>
        {selected.classification && <section className="analysis-card"><div className="analysis-heading"><div><p className="eyebrow">Deterministic classification</p><h3>{selected.classification.disposition.replaceAll('_', ' ')}</h3></div><strong>{(selected.classification.confidence_bp / 100).toFixed(0)}%</strong></div><p>{selected.classification.explanation}</p>{selected.classification.auto_rejected && <div className="exclusion-warning">Approved deterministic exclusion · not an AI recommendation</div>}<ul>{selected.classification.evidence.map((evidence, index) => <li key={`${evidence.rule_code}-${index}`}><strong>{evidence.rule_code}</strong><span>“{evidence.keyword}” in {evidence.source_field}</span></li>)}</ul></section>}
        {selected.score && <section className="analysis-card"><div className="analysis-heading"><div><p className="eyebrow">Explainable score</p><h3>{(selected.score.total_score_bp / 100).toFixed(0)} / 100</h3></div><strong>{selected.score.needs_review ? 'Review' : 'Ready'}</strong></div><p>{selected.score.explanation}</p><div className="score-breakdown">{selected.score.dimensions.map(dimension => <div key={dimension.code}><span>{dimension.code.replaceAll('_', ' ')}</span><progress max="10000" value={dimension.raw_score_bp} /><strong>{(dimension.raw_score_bp / 100).toFixed(0)}</strong><small>{dimension.explanation}</small></div>)}</div></section>}
        <section className="ai-separation"><strong>AI recommendation</strong><span>Belum dijalankan. Hasil di atas berasal dari deterministic rules dan scoring.</span></section>
        <a className="official-link" href={selected.canonical_url} target="_blank" rel="noreferrer">Open official source ↗</a><p className="detail-note">Verify all procurement details on the official source before making a decision.</p>
      </section></div>}
    </div>
  )
}
