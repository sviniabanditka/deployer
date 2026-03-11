import client from './client'

export interface App {
  id: string
  name: string
  slug?: string
  status: string
  domain?: string
  envVars?: Record<string, string>
  createdAt: string
  updatedAt: string
}

export interface Deployment {
  id: string
  appId: string
  version: string
  status: string
  imageTag?: string
  buildLog?: string
  isPreview?: boolean
  previewUrl?: string
  prNumber?: number
  prUrl?: string
  isCurrent?: boolean
  createdAt: string
}

export interface AppStats {
  cpuPercent: number
  memoryUsedMb: number
  memoryTotalMb: number
  networkInKb: number
  networkOutKb: number
}

export function fetchApps() {
  return client.get<App[]>('/apps')
}

export function fetchApp(id: string) {
  return client.get<App>(`/apps/${id}`)
}

export function createApp(payload: { name: string }) {
  return client.post<App>('/apps', payload)
}

export function deleteApp(id: string) {
  return client.delete(`/apps/${id}`)
}

export function deployApp(id: string, file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return client.post(`/apps/${id}/deploy`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function updateEnvVars(id: string, envVars: Record<string, string>) {
  return client.put(`/apps/${id}/env`, { envVars })
}

export function stopApp(id: string) {
  return client.post(`/apps/${id}/stop`)
}

export function startApp(id: string) {
  return client.post(`/apps/${id}/start`)
}

export function getAppStats(id: string) {
  return client.get<AppStats>(`/apps/${id}/stats`)
}

export function getDeployments(id: string) {
  return client.get<Deployment[]>(`/apps/${id}/deployments`)
}

export function rollbackApp(appId: string, deploymentId: string) {
  return client.post<void>(`/apps/${appId}/rollback`, { deploymentId })
}
