import client from './client'

export interface GitConnection {
  id: string
  appId: string
  provider: 'github' | 'gitlab'
  repoUrl: string
  branch: string
  autoDeploy: boolean
  connectedAt: string
}

export interface ConnectRepoPayload {
  provider: 'github' | 'gitlab'
  repoUrl: string
  branch: string
  accessToken: string
}

export function connectRepo(appId: string, payload: ConnectRepoPayload) {
  return client.post<GitConnection>(`/apps/${appId}/git`, payload)
}

export function disconnectRepo(appId: string) {
  return client.delete(`/apps/${appId}/git`)
}

export function getGitConnection(appId: string) {
  return client.get<GitConnection>(`/apps/${appId}/git`)
}

export function toggleAutoDeploy(appId: string, enabled: boolean) {
  return client.patch<GitConnection>(`/apps/${appId}/git`, { autoDeploy: enabled })
}
