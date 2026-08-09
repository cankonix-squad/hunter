import React from 'react'
import ReactDOM from 'react-dom/client'
import { AuthProvider, useAuth } from './auth'
import { Dashboard } from './dashboard'
import { LoginPage } from './login'
import './styles.css'

function Application() {
  const auth = useAuth()

  if (auth.status === 'loading') {
    return <div className="boot" role="status"><span className="spinner" />Restoring secure session…</div>
  }

  return auth.status === 'authenticated' ? <Dashboard /> : <LoginPage />
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AuthProvider><Application /></AuthProvider>
  </React.StrictMode>,
)
