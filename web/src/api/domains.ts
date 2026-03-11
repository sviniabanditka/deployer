import client from './client'

export interface CustomDomain {
  id: string
  appId: string
  domain: string
  verificationToken: string
  status: 'pending' | 'verified' | 'failed'
  createdAt: string
}

export function addDomain(appId: string, domain: string) {
  return client.post<CustomDomain>(`/apps/${appId}/domains`, { domain })
}

export function listDomains(appId: string) {
  return client.get<CustomDomain[]>(`/apps/${appId}/domains`)
}

export function verifyDomain(appId: string, domainId: string) {
  return client.post<void>(`/apps/${appId}/domains/${domainId}/verify`)
}

export function removeDomain(appId: string, domainId: string) {
  return client.delete(`/apps/${appId}/domains/${domainId}`)
}
