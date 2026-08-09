import { type FormEvent, useState } from 'react'
import { APIError } from './api'
import { useAuth } from './auth'

export function LoginPage() {
  const auth = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setSubmitting(true)
    try { await auth.login(email, password) }
    catch (reason) {
      setError(reason instanceof APIError && reason.status === 401
        ? 'Email atau password tidak cocok.'
        : 'Login tidak tersedia. Periksa koneksi dan coba lagi.')
      setSubmitting(false)
    }
  }

  return (
    <main className="login-page">
      <section className="login-intro">
        <div className="brand brand-large"><span>CX</span><strong>CANKONIX</strong></div>
        <div><p className="eyebrow">Opportunity intelligence</p><h1>See the signal.<br />Own the decision.</h1></div>
        <p className="login-note">Internal, auditable decision support for public procurement opportunities.</p>
      </section>
      <section className="login-panel">
        <form onSubmit={submit} aria-label="Login">
          <p className="eyebrow">Secure workspace</p>
          <h2>Sign in</h2>
          <p className="form-copy">Use your CANKONIX workspace credentials.</p>
          <label>Email<input autoFocus required type="email" autoComplete="username" value={email} onChange={e => setEmail(e.target.value)} placeholder="name@cankonix.com" /></label>
          <label>Password<input required minLength={8} type="password" autoComplete="current-password" value={password} onChange={e => setPassword(e.target.value)} /></label>
          {error && <div className="form-error" role="alert">{error}</div>}
          <button disabled={submitting} type="submit">{submitting ? 'Signing in…' : 'Sign in'}</button>
          <small>Protected by rotating sessions and role-based access control.</small>
        </form>
      </section>
    </main>
  )
}
