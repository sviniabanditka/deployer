import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as dbApi from '../api/databases'
import type { ManagedDatabase, DatabaseBackup } from '../api/databases'

export const useDatabasesStore = defineStore('databases', () => {
  const databases = ref<ManagedDatabase[]>([])
  const currentDatabase = ref<ManagedDatabase | null>(null)
  const backups = ref<DatabaseBackup[]>([])
  const loading = ref(false)

  async function fetchDatabases() {
    loading.value = true
    try {
      const { data } = await dbApi.fetchDatabases()
      databases.value = data
    } finally {
      loading.value = false
    }
  }

  async function fetchDatabase(id: string) {
    loading.value = true
    try {
      const { data } = await dbApi.fetchDatabase(id)
      currentDatabase.value = data
    } finally {
      loading.value = false
    }
  }

  async function createDatabase(payload: { name: string; engine: string; version?: string; appId?: string }) {
    const { data } = await dbApi.createDatabase(payload)
    databases.value.push(data)
    return data
  }

  async function deleteDatabase(id: string) {
    await dbApi.deleteDatabase(id)
    databases.value = databases.value.filter((d) => d.id !== id)
    if (currentDatabase.value?.id === id) {
      currentDatabase.value = null
    }
  }

  async function stopDatabase(id: string) {
    await dbApi.stopDatabase(id)
    await fetchDatabase(id)
  }

  async function startDatabase(id: string) {
    await dbApi.startDatabase(id)
    await fetchDatabase(id)
  }

  async function linkToApp(id: string, appId: string) {
    await dbApi.linkDatabase(id, appId)
    await fetchDatabase(id)
  }

  async function unlinkFromApp(id: string) {
    await dbApi.unlinkDatabase(id)
    await fetchDatabase(id)
  }

  async function createBackup(id: string) {
    const { data } = await dbApi.createBackup(id)
    backups.value.unshift(data)
    return data
  }

  async function fetchBackups(id: string) {
    try {
      const { data } = await dbApi.fetchBackups(id)
      backups.value = data
    } catch {
      backups.value = []
    }
  }

  async function restoreBackup(dbId: string, backupId: string) {
    await dbApi.restoreBackup(dbId, backupId)
    await fetchDatabase(dbId)
  }

  return {
    databases,
    currentDatabase,
    backups,
    loading,
    fetchDatabases,
    fetchDatabase,
    createDatabase,
    deleteDatabase,
    stopDatabase,
    startDatabase,
    linkToApp,
    unlinkFromApp,
    createBackup,
    fetchBackups,
    restoreBackup,
  }
})
