import { useEffect, useState } from 'react'
import { useAuth } from './auth'

interface Source { id: string; name: string; status: string; health_score: number }

export function Dashboard() {
  const auth = useAuth()
  const [sources, setSources] = useState<Source[]>([])
  const [degraded, setDegraded] = useState(false)

  useEffect(() => {
    let active = true
    auth.request<{ items: Source[] }>('/api/v1/sources')
      .then(result => { if (active) setSources(result.items) })
      .catch(() => { if (active) setDegraded(true) })
    return () => { active = false }
  }, [auth.request])

  const activeSources = sources.filter(source => source.status === 'active').length
  const metrics = [
    ['Priority opportunities', '—', 'Canonical feed is next'],
    ['Deadline < 7 hari', '—', 'Schedule pipeline pending'],
    ['Active sources', `${activeSources} / ${sources.length}`, degraded ? 'Source API degraded' : 'Registry connected'],
    ['Needs review', '—', 'AI enrichment pending'],
  ]

  return (
    <div className="shell">
      <aside>
        <div className="brand"><span>CX</span><strong>CANKONIX</strong></div>
        <nav aria-label="Navigasi utama"><a className="active" href="#overview">Overview</a><a href="#opportunities">Opportunities</a><a href="#sources">Sources</a><a href="#operations">Operations</a></nav>
        <div className="profile"><strong>{auth.session?.user.name}</strong><span>{auth.session?.user.roles.join(' · ')}</span><button className="text-button" onClick={() => void auth.logout()}>Sign out</button></div>
        <p className="build">Opportunity Intelligence<br />Collector foundation · v0.2.0</p>
      </aside>
      <main>
        <header><div><p className="eyebrow">Decision support</p><h1>Opportunity overview</h1></div><div className="status"><i /> System ready</div></header>
        <section className="metrics" aria-label="Ringkasan">{metrics.map(([label, value, note]) => <article key={label}><p>{label}</p><strong>{value}</strong><small>{note}</small></article>)}</section>
        <section className="empty"><div className="mark">03</div><p className="eyebrow">Collector online</p><h2>Normalize the first opportunity.</h2><p>The fixture pipeline is collecting idempotent source snapshots. Canonical package, lifecycle, and change history come next.</p><button type="button">View source registry</button></section>
      </main>
    </div>
  )
}
