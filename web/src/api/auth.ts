import client from './client'

export interface LoginPayload {
  email: string
  password: string
}

export interface RegisterPayload {
  name: string
  email: string
  password: string
}

export interface AuthResponse {
  user: { id: string; name: string; email: string }
  accessToken: string
  refreshToken: string
}

export interface LoginResponse extends AuthResponse {
  requires2FA?: boolean
  tempToken?: string
}

export interface TwoFactorSetupResponse {
  secret: string
  qrCodeUrl: string
}

export interface ChangePasswordPayload {
  currentPassword: string
  newPassword: string
}

export function login(payload: LoginPayload) {
  return client.post<LoginResponse>('/auth/login', payload)
}

export function register(payload: RegisterPayload) {
  return client.post<AuthResponse>('/auth/register', payload)
}

export function refreshToken(token: string) {
  return client.post<AuthResponse>('/auth/refresh', { refreshToken: token })
}

export function logout() {
  return client.post('/auth/logout')
}

export function enable2FA() {
  return client.post<TwoFactorSetupResponse>('/auth/2fa/enable')
}

export function verify2FA(code: string) {
  return client.post('/auth/2fa/verify', { code })
}

export function disable2FA(code: string) {
  return client.post('/auth/2fa/disable', { code })
}

export function validate2FA(tempToken: string, code: string) {
  return client.post<AuthResponse>('/auth/2fa/validate', { tempToken, code })
}

export function changePassword(payload: ChangePasswordPayload) {
  return client.post('/auth/change-password', payload)
}

export async function exportData(): Promise<Blob> {
  const response = await client.get('/account/export', { responseType: 'blob' })
  return response.data
}

export function deleteAccount(password: string) {
  return client.post('/account/delete', { password })
}
