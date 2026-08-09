import type { TokenResponse } from './session'

interface APIErrorBody { error?: { code?: string; message?: string } }

export class APIError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message)
    this.name = 'APIError'
  }
}

async function parse<T>(response: Response): Promise<T> {
  if (response.status === 204) return undefined as T
  const body = await response.json().catch(() => ({})) as T & APIErrorBody
  if (!response.ok) {
    throw new APIError(response.status, body.error?.code ?? 'request_failed', body.error?.message ?? 'Request failed')
  }
  return body
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  return parse<T>(response)
}

export const authAPI = {
  login: (email: string, password: string) => post<TokenResponse>('/api/v1/auth/login', { email, password }),
  refresh: (refreshToken: string) => post<TokenResponse>('/api/v1/auth/refresh', { refresh_token: refreshToken }),
  logout: (refreshToken: string) => post<void>('/api/v1/auth/logout', { refresh_token: refreshToken }),
}

export async function authorized<T>(path: string, accessToken: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: { ...init.headers, Authorization: `Bearer ${accessToken}` },
  })
  return parse<T>(response)
}
