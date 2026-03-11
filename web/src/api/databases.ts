import client from './client'

export interface ManagedDatabase {
  id: string
  userId: string
  appId?: string
  name: string
  engine: string
  version: string
  status: string
  host: string
  port: number
  dbName: string
  username: string
  password: string
  connectionUrl: string
  memoryLimit: number
  storageUsed: number
  storageLimit: number
  createdAt: string
  updatedAt: string
}

export interface DatabaseBackup {
  id: string
  databaseId: string
  filePath: string
  fileSize: number
  status: string
  createdAt: string
}

export function fetchDatabases() {
  return client.get<ManagedDatabase[]>('/databases')
}

export function fetchDatabase(id: string) {
  return client.get<ManagedDatabase>(`/databases/${id}`)
}

export function createDatabase(data: { name: string; engine: string; version?: string; appId?: string }) {
  return client.post<ManagedDatabase>('/databases', data)
}

export function deleteDatabase(id: string) {
  return client.delete(`/databases/${id}`)
}

export function stopDatabase(id: string) {
  return client.post(`/databases/${id}/stop`)
}

export function startDatabase(id: string) {
  return client.post(`/databases/${id}/start`)
}

export function linkDatabase(id: string, appId: string) {
  return client.post(`/databases/${id}/link`, { appId })
}

export function unlinkDatabase(id: string) {
  return client.post(`/databases/${id}/unlink`)
}

export function createBackup(id: string) {
  return client.post<DatabaseBackup>(`/databases/${id}/backups`)
}

export function fetchBackups(id: string) {
  return client.get<DatabaseBackup[]>(`/databases/${id}/backups`)
}

export function restoreBackup(dbId: string, backupId: string) {
  return client.post(`/databases/${dbId}/backups/${backupId}/restore`)
}

export function startWebUI(id: string) {
  return client.post<{ url: string }>(`/databases/${id}/webui/start`)
}

export function stopWebUI(id: string) {
  return client.post<void>(`/databases/${id}/webui/stop`)
}
