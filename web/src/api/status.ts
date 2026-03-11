import client from './client'

export interface ComponentStatus {
  name: string
  status: 'operational' | 'degraded' | 'outage'
  latency: number
  message?: string
}

export interface SystemStatus {
  overall: 'operational' | 'degraded' | 'outage'
  components: ComponentStatus[]
  updatedAt: string
}

export interface Incident {
  id: string
  component: string
  description: string
  severity: 'minor' | 'major' | 'critical'
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved'
  startedAt: string
  resolvedAt?: string
}

export function fetchStatus() {
  return client.get<SystemStatus>('/status')
}

export function fetchIncidents() {
  return client.get<Incident[]>('/status/incidents')
}
