import React from 'react'
import ReactDOM from 'react-dom/client'
import './styles.css'

const metrics = [
  ['Priority opportunities', '—', 'Menunggu sumber pertama'],
  ['Deadline < 7 hari', '—', 'Belum ada data aktif'],
  ['Source health', '0 / 0', 'Registry belum dikonfigurasi'],
  ['Needs review', '—', 'AI belum diaktifkan'],
]

function App() {
  return (
    <div className="shell">
      <aside>
        <div className="brand"><span>CX</span><strong>CANKONIX</strong></div>
        <nav aria-label="Navigasi utama">
          <a className="active" href="#overview">Overview</a>
          <a href="#opportunities">Opportunities</a>
          <a href="#sources">Sources</a>
          <a href="#operations">Operations</a>
        </nav>
        <p className="build">Opportunity Intelligence<br />Foundation · v0.1.0</p>
      </aside>
      <main>
        <header><div><p className="eyebrow">Decision support</p><h1>Opportunity overview</h1></div><div className="status"><i /> System ready</div></header>
        <section className="metrics" aria-label="Ringkasan">
          {metrics.map(([label, value, note]) => <article key={label}><p>{label}</p><strong>{value}</strong><small>{note}</small></article>)}
        </section>
        <section className="empty">
          <div className="mark">01</div>
          <p className="eyebrow">Foundation online</p>
          <h2>Connect the first public source.</h2>
          <p>Platform baseline is ready for the first SPSE adapter vertical slice: discover, fingerprint, normalize, and review.</p>
          <button type="button">Configure source</button>
        </section>
      </main>
    </div>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(<React.StrictMode><App /></React.StrictMode>)

