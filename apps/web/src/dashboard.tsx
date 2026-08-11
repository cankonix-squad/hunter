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
  filters: { q: string; status: string; source_id?: string; relevance: string; min_score_bp?: number; review_only: boolean; shortlisted?: string }
}
interface TeamMember { user_id: string; email: string; name: string; roles: string[] }
interface Assignment { opportunity_id: string; assignee_user_id: string; assignee_name: string; assignee_email: string; assigned_by: string; assigned_at: string; updated_at: string }
interface OpportunityComment { id: string; opportunity_id: string; author_user_id: string; author_name: string; body: string; created_at: string }
interface ShortlistState { opportunity_id: string; shortlisted: boolean; shortlisted_by?: string; actor_name?: string; updated_at?: string }
interface HumanDecision { opportunity_id: string; decision: 'pursue' | 'pass'; reason: string; decided_by: string; decider_name: string; decided_at: string; updated_at: string }
interface NextAction { opportunity_id: string; action: string; due_at: string; updated_by: string; updater_name: string; created_at: string; updated_at: string }
interface ReadinessChecklist { opportunity_id: string; version: string; completed: number; total: number; items: Array<{ code: string; label: string; completed: boolean; note: string; updater_name?: string; updated_at?: string }> }
interface HumanOverride { opportunity_id:string; recommendation?:string; score_bp?:number; reason:string; overridden_by:string; overrider_name:string; created_at:string; updated_at:string }
interface ActivityEvent { id:number; action:string; actor_name?:string; metadata:Record<string,unknown>; occurred_at:string }
interface PipelineAnalytics { total:number; active:number; shortlisted:number; assigned:number; pursue:number; pass:number; undecided:number; overdue_actions:number; actions_due_seven_days:number }
interface AlertRule { id:string; name:string; enabled:boolean; event_types:string[]; filters:{source_id?:string;relevance:string;min_score_bp?:number}; created_at:string; updated_at:string }
interface AlertEvent { id:number; rule_id:string; rule_name:string; opportunity_id:string; package_identity:string; title:string; event_type:string; changed_fields:string[]; delivery_status:string; created_at:string }
interface NotificationPreferences { quiet_hours_enabled:boolean; quiet_hours_start:string; quiet_hours_end:string; timezone:string; updated_at?:string }
interface DeadLetter { id:number; alert_event_id:number; channel:string; destination:string; status:string; attempts:number; replay_count:number; safe_error?:string; response_code?:number; event_type:string; opportunity_id:string; package_identity:string; title:string; updated_at:string; last_replayed_at?:string }

const money = new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 })
const date = new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Jakarta' })
const jakartaInput = (value: string) => new Date(new Date(value).getTime() + 7 * 3600000).toISOString().slice(0, 16)

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
  const [shortlistOnly, setShortlistOnly] = useState(false)
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState<Opportunity | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [sources, setSources] = useState<SourceHealth[]>([])
  const [sourceError, setSourceError] = useState('')
  const [savedViews, setSavedViews] = useState<SavedView[]>([])
  const [viewName, setViewName] = useState('')
  const [viewError, setViewError] = useState('')
  const [team, setTeam] = useState<TeamMember[]>([])
  const [assignment, setAssignment] = useState<Assignment | null>(null)
  const [assignmentError, setAssignmentError] = useState('')
  const [comments, setComments] = useState<OpportunityComment[]>([])
  const [commentInput, setCommentInput] = useState('')
  const [commentError, setCommentError] = useState('')
  const [shortlistState, setShortlistState] = useState<ShortlistState | null>(null)
  const [decision, setDecision] = useState<HumanDecision | null>(null)
  const [decisionChoice, setDecisionChoice] = useState('')
  const [decisionReason, setDecisionReason] = useState('')
  const [decisionError, setDecisionError] = useState('')
  const [nextAction, setNextAction] = useState<NextAction | null>(null)
  const [nextActionInput, setNextActionInput] = useState('')
  const [nextActionDue, setNextActionDue] = useState('')
  const [nextActionError, setNextActionError] = useState('')
  const [readiness, setReadiness] = useState<ReadinessChecklist | null>(null)
  const [readinessError, setReadinessError] = useState('')
  const [humanOverride,setHumanOverride]=useState<HumanOverride|null>(null);const [overrideRecommendation,setOverrideRecommendation]=useState('');const [overrideScore,setOverrideScore]=useState('');const [overrideReason,setOverrideReason]=useState('');const [overrideError,setOverrideError]=useState('')
  const [activity,setActivity]=useState<ActivityEvent[]>([])
  const [pipeline,setPipeline]=useState<PipelineAnalytics>({total:0,active:0,shortlisted:0,assigned:0,pursue:0,pass:0,undecided:0,overdue_actions:0,actions_due_seven_days:0})
  const [alertRules,setAlertRules]=useState<AlertRule[]>([]);const [alertEvents,setAlertEvents]=useState<AlertEvent[]>([]);const [alertName,setAlertName]=useState('');const [alertSource,setAlertSource]=useState('');const [alertRelevance,setAlertRelevance]=useState('all');const [alertMinScore,setAlertMinScore]=useState('');const [alertNew,setAlertNew]=useState(true);const [alertChanges,setAlertChanges]=useState(true);const [alertError,setAlertError]=useState('')
  const [quietHours,setQuietHours]=useState<NotificationPreferences>({quiet_hours_enabled:false,quiet_hours_start:'22:00',quiet_hours_end:'07:00',timezone:'Asia/Jakarta'})
  const [deadLetters,setDeadLetters]=useState<DeadLetter[]>([]);const [deadLetterError,setDeadLetterError]=useState('')

  const load = useCallback(() => {
    setLoading(true)
    setError('')
    const params = new URLSearchParams({ q: search, status, relevance, page: String(page), page_size: '20' })
    if (sourceID) params.set('source_id', sourceID)
    if (shortlistOnly) params.set('shortlisted', 'true')
    if (minimumScore) params.set('min_score', String(Math.round(Number(minimumScore) * 100)))
    const path = reviewOnly ? '/api/v1/opportunities/review' : '/api/v1/opportunities'
    auth.request<OpportunityPage>(`${path}?${params}`)
      .then(setResult)
      .catch(() => setError('Feed peluang belum dapat dimuat. Coba beberapa saat lagi.'))
      .finally(() => setLoading(false))
  }, [auth.request, minimumScore, page, relevance, reviewOnly, search, shortlistOnly, sourceID, status])

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

  useEffect(() => { auth.request<{ items: TeamMember[] }>('/api/v1/team').then(body => setTeam(body.items)).catch(() => setTeam([])) }, [auth.request])
  useEffect(() => { auth.request<PipelineAnalytics>('/api/v1/analytics/pipeline').then(setPipeline).catch(() => setPipeline({total:0,active:0,shortlisted:0,assigned:0,pursue:0,pass:0,undecided:0,overdue_actions:0,actions_due_seven_days:0})) }, [auth.request])
  const loadAlerts=useCallback(()=>{Promise.all([auth.request<{items:AlertRule[]}>('/api/v1/alert-rules'),auth.request<{items:AlertEvent[]}>('/api/v1/alerts'),auth.request<NotificationPreferences>('/api/v1/notification-preferences')]).then(([rules,events,preferences])=>{setAlertRules(rules.items);setAlertEvents(events.items);setQuietHours({...preferences,quiet_hours_start:preferences.quiet_hours_start.slice(0,5),quiet_hours_end:preferences.quiet_hours_end.slice(0,5)});setAlertError('')}).catch(()=>setAlertError('Alert rules belum dapat dimuat.'))},[auth.request]);useEffect(loadAlerts,[loadAlerts])

  async function createAlertRule(event:FormEvent){event.preventDefault();const eventTypes=[...(alertNew?['opportunity.new']:[]),...(alertChanges?['opportunity.material_change']:[])];if(!alertName.trim()||eventTypes.length===0){setAlertError('Nama dan minimal satu event wajib dipilih.');return}try{const item=await auth.request<AlertRule>('/api/v1/alert-rules',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:alertName.trim(),event_types:eventTypes,filters:{source_id:alertSource||undefined,relevance:alertRelevance,min_score_bp:alertMinScore?Math.round(Number(alertMinScore)*100):undefined}})});setAlertRules(items=>[...items,item].sort((a,b)=>a.name.localeCompare(b.name)));setAlertName('');setAlertError('')}catch{setAlertError('Alert rule gagal disimpan. Gunakan nama unik dan filter valid.')}}
  async function deleteAlertRule(id:string){try{await auth.request(`/api/v1/alert-rules/${id}`,{method:'DELETE'});setAlertRules(items=>items.filter(item=>item.id!==id));setAlertError('')}catch{setAlertError('Alert rule gagal dihapus.')}}
  async function saveQuietHours(event:FormEvent){event.preventDefault();try{const item=await auth.request<NotificationPreferences>('/api/v1/notification-preferences',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(quietHours)});setQuietHours({...item,quiet_hours_start:item.quiet_hours_start.slice(0,5),quiet_hours_end:item.quiet_hours_end.slice(0,5)});setAlertError('')}catch{setAlertError('Quiet hours gagal disimpan.')}}
  const loadDeadLetters=useCallback(()=>{if(!auth.session?.user.roles.some(role=>['administrator','operator'].includes(role)))return;auth.request<{items:DeadLetter[]}>('/api/v1/operations/dead-letters').then(body=>{setDeadLetters(body.items);setDeadLetterError('')}).catch(()=>setDeadLetterError('Dead-letter queue belum dapat dimuat.'))},[auth.request,auth.session?.user.roles]);useEffect(loadDeadLetters,[loadDeadLetters])
  async function replayDeadLetter(id:number){try{await auth.request<DeadLetter>(`/api/v1/operations/dead-letters/${id}/replay`,{method:'POST'});setDeadLetters(items=>items.filter(item=>item.id!==id));setDeadLetterError('')}catch{setDeadLetterError('Delivery gagal diantrikan ulang.')}}

  function applySavedView(id: string) {
    const view = savedViews.find(item => item.id === id)
    if (!view) return
    setSearchInput(view.filters.q); setSearch(view.filters.q); setStatus(view.filters.status)
    setSourceID(view.filters.source_id ?? '')
    setRelevance(view.filters.relevance); setMinimumScore(view.filters.min_score_bp == null ? '' : String(view.filters.min_score_bp / 100))
    setReviewOnly(view.filters.review_only); setPage(1); setViewError('')
    setShortlistOnly(view.filters.shortlisted === 'true')
  }

  async function saveCurrentView(event: FormEvent) {
    event.preventDefault()
    const name = viewName.trim()
    if (!name) { setViewError('Nama saved view wajib diisi.'); return }
    try {
      const item = await auth.request<SavedView>('/api/v1/saved-views', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name, filters: { q: search, status, source_id: sourceID || undefined, relevance, min_score_bp: minimumScore ? Math.round(Number(minimumScore) * 100) : undefined, review_only: reviewOnly, shortlisted: shortlistOnly ? 'true' : 'all' } }) })
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
    setAssignmentError('')
    try {
      const [detail, owner, discussion, shortlist, humanDecision, action, checklist, override, events] = await Promise.all([auth.request<Opportunity>(`/api/v1/opportunities/${id}`), auth.request<{ assignment: Assignment | null }>(`/api/v1/opportunities/${id}/assignment`), auth.request<{ items: OpportunityComment[] }>(`/api/v1/opportunities/${id}/comments`), auth.request<ShortlistState>(`/api/v1/opportunities/${id}/shortlist`), auth.request<{ decision: HumanDecision | null }>(`/api/v1/opportunities/${id}/decision`), auth.request<{ next_action: NextAction | null }>(`/api/v1/opportunities/${id}/next-action`), auth.request<ReadinessChecklist>(`/api/v1/opportunities/${id}/readiness`),auth.request<{override:HumanOverride|null}>(`/api/v1/opportunities/${id}/override`),auth.request<{items:ActivityEvent[]}>(`/api/v1/opportunities/${id}/activity`)])
      setSelected(detail); setAssignment(owner.assignment); setComments(discussion.items); setShortlistState(shortlist); setDecision(humanDecision.decision); setDecisionChoice(humanDecision.decision?.decision ?? ''); setDecisionReason(humanDecision.decision?.reason ?? ''); setDecisionError(''); setNextAction(action.next_action); setNextActionInput(action.next_action?.action ?? ''); setNextActionDue(action.next_action ? jakartaInput(action.next_action.due_at) : ''); setNextActionError(''); setReadiness(checklist); setReadinessError('');setHumanOverride(override.override);setOverrideRecommendation(override.override?.recommendation??'');setOverrideScore(override.override?.score_bp==null?'':String(override.override.score_bp/100));setOverrideReason(override.override?.reason??'');setOverrideError('');setActivity(events.items); setCommentInput(''); setCommentError('')
    }
    catch { setError('Detail peluang belum dapat dimuat.') }
  }

  async function assignOwner(userID: string) {
    if (!selected) return
    try {
      const body = await auth.request<{ assignment: Assignment | null }>(`/api/v1/opportunities/${selected.id}/assignment`, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ assignee_user_id: userID || null }) })
      setAssignment(body.assignment); setAssignmentError('')
    } catch { setAssignmentError('Assignment gagal diperbarui.') }
  }

  async function submitComment(event: FormEvent) {
    event.preventDefault()
    if (!selected || !commentInput.trim()) { setCommentError('Komentar wajib diisi.'); return }
    try {
      const item = await auth.request<OpportunityComment>(`/api/v1/opportunities/${selected.id}/comments`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ body: commentInput.trim() }) })
      setComments(items => [...items,item]); setCommentInput(''); setCommentError('')
    } catch { setCommentError('Komentar gagal disimpan.') }
  }

  async function toggleShortlist() {
    if (!selected || !shortlistState) return
    try { setShortlistState(await auth.request<ShortlistState>(`/api/v1/opportunities/${selected.id}/shortlist`, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ shortlisted: !shortlistState.shortlisted }) })) }
    catch { setError('Shortlist gagal diperbarui.') }
  }

  async function submitDecision(event: FormEvent) {
    event.preventDefault()
    if (!selected || !decisionChoice || decisionReason.trim().length < 3) { setDecisionError('Pilih keputusan dan isi alasan minimal 3 karakter.'); return }
    try {
      const body = await auth.request<{ decision: HumanDecision }>(`/api/v1/opportunities/${selected.id}/decision`, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ decision: decisionChoice, reason: decisionReason.trim() }) })
      setDecision(body.decision); setDecisionReason(body.decision.reason); setDecisionError('')
    } catch { setDecisionError('Keputusan gagal disimpan.') }
  }

  async function submitNextAction(event: FormEvent) {
    event.preventDefault(); if (!selected || nextActionInput.trim().length < 3 || !nextActionDue) { setNextActionError('Action dan deadline wajib diisi.'); return }
    try { const body=await auth.request<{next_action:NextAction}>(`/api/v1/opportunities/${selected.id}/next-action`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({action:nextActionInput.trim(),due_at:new Date(`${nextActionDue}:00+07:00`).toISOString()})});setNextAction(body.next_action);setNextActionError('') } catch { setNextActionError('Next action gagal disimpan.') }
  }
  async function clearNextAction() { if(!selected)return;try{await auth.request(`/api/v1/opportunities/${selected.id}/next-action`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({action:null,due_at:null})});setNextAction(null);setNextActionInput('');setNextActionDue('');setNextActionError('')}catch{setNextActionError('Next action gagal dihapus.')} }
  async function updateReadiness(code:string,completed:boolean,note:string){if(!selected)return;try{setReadiness(await auth.request<ReadinessChecklist>(`/api/v1/opportunities/${selected.id}/readiness`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({code,completed,note})}));setReadinessError('')}catch{setReadinessError('Readiness checklist gagal diperbarui.')}}
  async function submitOverride(event:FormEvent){event.preventDefault();if(!selected||(!overrideRecommendation&&!overrideScore)||overrideReason.trim().length<3){setOverrideError('Isi recommendation atau score dan alasan minimal 3 karakter.');return}try{const body=await auth.request<{override:HumanOverride}>(`/api/v1/opportunities/${selected.id}/override`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({recommendation:overrideRecommendation||null,score_bp:overrideScore?Math.round(Number(overrideScore)*100):null,reason:overrideReason.trim(),clear:false})});setHumanOverride(body.override);setOverrideError('')}catch{setOverrideError('Override gagal disimpan.')}}
  async function clearOverride(){if(!selected)return;try{await auth.request(`/api/v1/opportunities/${selected.id}/override`,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({recommendation:null,score_bp:null,reason:'',clear:true})});setHumanOverride(null);setOverrideRecommendation('');setOverrideScore('');setOverrideReason('');setOverrideError('')}catch{setOverrideError('Override gagal dihapus.')}}

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

        <section className="pipeline-analytics" id="overview" aria-labelledby="pipeline-heading">
          <div><p className="eyebrow">Portfolio overview</p><h2 id="pipeline-heading">Pipeline analytics</h2><small>{pipeline.active} active of {pipeline.total} total opportunities</small></div>
          <dl><div><dt>Pursue</dt><dd>{pipeline.pursue}</dd></div><div><dt>Pass</dt><dd>{pipeline.pass}</dd></div><div><dt>Undecided</dt><dd>{pipeline.undecided}</dd></div><div><dt>Shortlisted</dt><dd>{pipeline.shortlisted}</dd></div><div><dt>Assigned</dt><dd>{pipeline.assigned}</dd></div><div className={pipeline.overdue_actions > 0 ? 'risk' : ''}><dt>Overdue actions</dt><dd>{pipeline.overdue_actions}</dd><small>{pipeline.actions_due_seven_days} due in 7 days</small></div></dl>
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
            <label className="review-toggle"><input aria-label="Shortlist only" type="checkbox" checked={shortlistOnly} onChange={event => { setShortlistOnly(event.target.checked); setPage(1) }} />Shortlist</label>
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
          <form className="quiet-hours" onSubmit={saveQuietHours}><p className="eyebrow">Delivery schedule</p><label><input aria-label="Enable quiet hours" type="checkbox" checked={quietHours.quiet_hours_enabled} onChange={e=>setQuietHours({...quietHours,quiet_hours_enabled:e.target.checked})}/>Quiet hours</label><label>Start<input aria-label="Quiet hours start" type="time" value={quietHours.quiet_hours_start} onChange={e=>setQuietHours({...quietHours,quiet_hours_start:e.target.value})}/></label><label>End<input aria-label="Quiet hours end" type="time" value={quietHours.quiet_hours_end} onChange={e=>setQuietHours({...quietHours,quiet_hours_end:e.target.value})}/></label><select aria-label="Quiet hours timezone" value={quietHours.timezone} onChange={e=>setQuietHours({...quietHours,timezone:e.target.value})}><option value="Asia/Jakarta">Asia/Jakarta</option><option value="Asia/Makassar">Asia/Makassar</option><option value="Asia/Jayapura">Asia/Jayapura</option></select><button type="submit">Save quiet hours</button></form>
          <div className="alert-console"><div className="section-heading"><div><p className="eyebrow">Notifications</p><h2>Alert rules</h2></div><button onClick={loadAlerts}>Refresh</button></div><form onSubmit={createAlertRule}><input aria-label="Alert rule name" maxLength={80} value={alertName} onChange={e=>setAlertName(e.target.value)} placeholder="Rule name…"/><select aria-label="Alert source" value={alertSource} onChange={e=>setAlertSource(e.target.value)}><option value="">All sources</option>{sources.map(source=><option key={source.id} value={source.id}>{source.name}</option>)}</select><select aria-label="Alert relevance" value={alertRelevance} onChange={e=>setAlertRelevance(e.target.value)}><option value="all">All relevance</option><option value="relevant">Relevant</option><option value="needs_review">Needs review</option><option value="excluded">Excluded</option></select><input aria-label="Alert minimum score" type="number" min="0" max="100" value={alertMinScore} onChange={e=>setAlertMinScore(e.target.value)} placeholder="Min score"/><label><input type="checkbox" checked={alertNew} onChange={e=>setAlertNew(e.target.checked)}/>New</label><label><input type="checkbox" checked={alertChanges} onChange={e=>setAlertChanges(e.target.checked)}/>Material changes</label><button type="submit">Add rule</button></form>{alertError&&<small role="alert">{alertError}</small>}<div className="alert-layout"><div>{alertRules.length===0?<p>No alert rules yet.</p>:alertRules.map(rule=><article key={rule.id}><strong>{rule.name}</strong><span>{rule.event_types.map(type=>type.replace('opportunity.','').replaceAll('_',' ')).join(' · ')} · {rule.filters.relevance}</span><button aria-label={`Delete alert rule ${rule.name}`} onClick={()=>void deleteAlertRule(rule.id)}>Delete</button></article>)}</div><div><h3>Recent alert events</h3>{alertEvents.length===0?<p>No matched events yet.</p>:alertEvents.map(event=><article key={event.id}><strong>{event.title}</strong><span>{event.event_type.replace('opportunity.','').replaceAll('_',' ')} · {event.rule_name}</span><small>{event.changed_fields.join(' · ')||'Initial observation'} · {date.format(new Date(event.created_at))}</small></article>)}</div></div></div>
          {auth.session?.user.roles.some(role=>['administrator','operator'].includes(role))&&<div className="dead-letter-console"><div className="section-heading"><div><p className="eyebrow">Failure recovery</p><h2>Dead-letter queue</h2></div><button onClick={loadDeadLetters}>Refresh</button></div>{deadLetterError&&<small role="alert">{deadLetterError}</small>}{deadLetters.length===0?<p>No dead-letter deliveries.</p>:deadLetters.map(item=><article key={item.id}><div><strong>{item.title}</strong><span>{item.channel} · {item.destination} · {item.attempts} attempts · replayed {item.replay_count}×</span><small>{item.safe_error||'No safe error recorded'}</small></div><button onClick={()=>void replayDeadLetter(item.id)}>Replay</button></article>)}</div>}
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
      {selected && <div className="detail-backdrop" role="presentation" onMouseDown={() => { setSelected(null); setAssignment(null) }}><section className="detail-panel" role="dialog" aria-modal="true" aria-labelledby="detail-title" onMouseDown={event => event.stopPropagation()}>
        <button className="detail-close" aria-label="Tutup detail" onClick={() => setSelected(null)}>×</button><p className="eyebrow">{selected.package_identity}</p><h2 id="detail-title">{selected.title}</h2><span className={`lifecycle ${selected.lifecycle_status}`}>{selected.lifecycle_status.replaceAll('_', ' ')}</span>
        <dl><div><dt>Instansi</dt><dd>{selected.agency}</dd></div><div><dt>Jenis pengadaan</dt><dd>{selected.procurement_type}</dd></div><div><dt>Nilai paket</dt><dd>{selected.budget_minor == null ? 'Belum tersedia' : money.format(selected.budget_minor)}</dd></div><div><dt>Batas pendaftaran</dt><dd>{selected.registration_deadline ? date.format(new Date(selected.registration_deadline)) : 'Belum tersedia'}</dd></div><div><dt>Sumber</dt><dd>{selected.source_name}</dd></div><div><dt>Riwayat terdeteksi</dt><dd>{selected.version_count} versi · pertama {date.format(new Date(selected.first_seen_at))}</dd></div></dl>
        <section className="assignment-card"><p className="eyebrow">Opportunity owner</p><strong>{assignment?.assignee_name ?? 'Unassigned'}</strong><select aria-label="Assign opportunity owner" value={assignment?.assignee_user_id ?? ''} onChange={event => void assignOwner(event.target.value)} disabled={!auth.session?.user.roles.some(role => ['administrator','operator','business_development','bid_manager'].includes(role))}><option value="">Unassigned</option>{team.map(member => <option key={member.user_id} value={member.user_id}>{member.name}</option>)}</select>{assignmentError && <small role="alert">{assignmentError}</small>}</section>
        {shortlistState && <button className={`shortlist-button ${shortlistState.shortlisted ? 'selected' : ''}`} onClick={() => void toggleShortlist()}>{shortlistState.shortlisted ? '★ Shortlisted' : '☆ Add to shortlist'}</button>}
        <section className={`decision-card ${decision?.decision ?? ''}`}><p className="eyebrow">Human decision</p><h3>{decision ? decision.decision : 'Undecided'}</h3>{decision && <p>{decision.reason}<small>Decided by {decision.decider_name} · {date.format(new Date(decision.updated_at))}</small></p>}<form onSubmit={submitDecision}><select aria-label="Pursue or pass decision" value={decisionChoice} onChange={event => setDecisionChoice(event.target.value)}><option value="">Choose decision…</option><option value="pursue">Pursue</option><option value="pass">Pass</option></select><textarea aria-label="Decision reason" maxLength={2000} value={decisionReason} onChange={event => setDecisionReason(event.target.value)} placeholder="Required reason…" /><button type="submit" disabled={!auth.session?.user.roles.some(role => ['administrator','operator','business_development','bid_manager','executive'].includes(role))}>Save decision</button>{decisionError && <small role="alert">{decisionError}</small>}</form></section>
        <section className={`next-action-card ${nextAction && new Date(nextAction.due_at).getTime() < Date.now() ? 'overdue' : ''}`}><p className="eyebrow">Next action</p>{nextAction && <p><strong>{nextAction.action}</strong><span>Due {date.format(new Date(nextAction.due_at))} · updated by {nextAction.updater_name}</span></p>}<form onSubmit={submitNextAction}><input aria-label="Next action" maxLength={500} value={nextActionInput} onChange={event=>setNextActionInput(event.target.value)} placeholder="e.g. Validate technical scope"/><label>Deadline · Asia/Jakarta<input aria-label="Next action deadline" type="datetime-local" value={nextActionDue} onChange={event=>setNextActionDue(event.target.value)}/></label><div><button type="submit">Save action</button>{nextAction&&<button type="button" onClick={()=>void clearNextAction()}>Clear</button>}</div>{nextActionError&&<small role="alert">{nextActionError}</small>}</form></section>
        {readiness&&<section className="readiness-card"><div><p className="eyebrow">Readiness checklist · {readiness.version}</p><strong>{readiness.completed}/{readiness.total}</strong></div><progress max={readiness.total} value={readiness.completed}/>{readiness.items.map(item=><article key={item.code}><label><input type="checkbox" aria-label={item.label} checked={item.completed} onChange={event=>void updateReadiness(item.code,event.target.checked,item.note)}/><span>{item.label}</span></label><input aria-label={`Note for ${item.label}`} maxLength={1000} value={item.note} onChange={event=>setReadiness(current=>current?{...current,items:current.items.map(existing=>existing.code===item.code?{...existing,note:event.target.value}:existing)}:current)} placeholder="Optional evidence or note"/><button onClick={()=>void updateReadiness(item.code,item.completed,item.note)}>Save note</button></article>)}{readinessError&&<small role="alert">{readinessError}</small>}</section>}
        <section className="override-card"><p className="eyebrow">Human override · separate from model output</p>{humanOverride&&<p><strong>{humanOverride.recommendation?.replaceAll('_',' ')??'Recommendation unchanged'} · {humanOverride.score_bp==null?'Score unchanged':`${humanOverride.score_bp/100} score`}</strong><span>{humanOverride.reason} · {humanOverride.overrider_name}</span></p>}<form onSubmit={submitOverride}><select aria-label="Recommendation override" value={overrideRecommendation} onChange={e=>setOverrideRecommendation(e.target.value)}><option value="">Keep recommendation</option><option value="relevant">Relevant</option><option value="excluded">Excluded</option><option value="needs_review">Needs review</option></select><input aria-label="Score override" type="number" min="0" max="100" value={overrideScore} onChange={e=>setOverrideScore(e.target.value)} placeholder="Keep score"/><textarea aria-label="Override reason" maxLength={2000} value={overrideReason} onChange={e=>setOverrideReason(e.target.value)} placeholder="Mandatory reason…"/><div><button type="submit">Save override</button>{humanOverride&&<button type="button" onClick={()=>void clearOverride()}>Clear override</button>}</div>{overrideError&&<small role="alert">{overrideError}</small>}</form></section>
        <section className="activity-log"><p className="eyebrow">Audit activity</p><h3>{activity.length} events</h3>{activity.length===0?<p>No workflow activity yet.</p>:activity.map(event=><article key={event.id}><strong>{event.action.replaceAll('.',' · ').replaceAll('_',' ')}</strong><span>{event.actor_name||'System'} · {date.format(new Date(event.occurred_at))}</span>{Object.keys(event.metadata).length>0&&<code>{JSON.stringify(event.metadata)}</code>}</article>)}</section>
        {selected.classification && <section className="analysis-card"><div className="analysis-heading"><div><p className="eyebrow">Deterministic classification</p><h3>{selected.classification.disposition.replaceAll('_', ' ')}</h3></div><strong>{(selected.classification.confidence_bp / 100).toFixed(0)}%</strong></div><p>{selected.classification.explanation}</p>{selected.classification.auto_rejected && <div className="exclusion-warning">Approved deterministic exclusion · not an AI recommendation</div>}<ul>{selected.classification.evidence.map((evidence, index) => <li key={`${evidence.rule_code}-${index}`}><strong>{evidence.rule_code}</strong><span>“{evidence.keyword}” in {evidence.source_field}</span></li>)}</ul></section>}
        {selected.score && <section className="analysis-card"><div className="analysis-heading"><div><p className="eyebrow">Explainable score</p><h3>{(selected.score.total_score_bp / 100).toFixed(0)} / 100</h3></div><strong>{selected.score.needs_review ? 'Review' : 'Ready'}</strong></div><p>{selected.score.explanation}</p><div className="score-breakdown">{selected.score.dimensions.map(dimension => <div key={dimension.code}><span>{dimension.code.replaceAll('_', ' ')}</span><progress max="10000" value={dimension.raw_score_bp} /><strong>{(dimension.raw_score_bp / 100).toFixed(0)}</strong><small>{dimension.explanation}</small></div>)}</div></section>}
        <section className="timeline"><p className="eyebrow">Change timeline</p><h3>{selected.history?.length ?? 0} observations</h3>{selected.history?.map((version, index) => <article key={version.id}><i /><div><strong>{index === selected.history!.length - 1 ? 'Initial observation' : version.changed_fields.length ? version.changed_fields.map(field => field.replaceAll('_', ' ')).join(' · ') : 'Observed again'}</strong><time>{date.format(new Date(version.observed_at))}</time><span>{version.lifecycle_status.replaceAll('_', ' ')} · snapshot {version.snapshot_id.slice(0, 8)}</span></div></article>)}</section>
        <section className="comments"><p className="eyebrow">Internal discussion</p><h3>{comments.length} comments</h3>{comments.length === 0 ? <p>No internal comments yet.</p> : <div>{comments.map(comment => <article key={comment.id}><header><strong>{comment.author_name}</strong><time>{date.format(new Date(comment.created_at))}</time></header><p>{comment.body}</p></article>)}</div>}<form onSubmit={submitComment}><textarea aria-label="Internal comment" maxLength={4000} value={commentInput} onChange={event => setCommentInput(event.target.value)} placeholder="Add context for the team…" /><button type="submit">Add comment</button>{commentError && <small role="alert">{commentError}</small>}</form></section>
        <section className="ai-separation"><strong>AI recommendation</strong><span>Belum dijalankan. Hasil di atas berasal dari deterministic rules dan scoring.</span></section>
        <a className="official-link" href={selected.canonical_url} target="_blank" rel="noreferrer">Open official source ↗</a><p className="detail-note">Verify all procurement details on the official source before making a decision.</p>
      </section></div>}
    </div>
  )
}
