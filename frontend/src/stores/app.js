import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAppStore = defineStore('app', () => {
  const accounts = ref([])
  const activeAccountId = ref(null)
  const zones = ref([])
  const certList = ref([])

  const connected = computed(() => accounts.value.length > 0)
  const activeAccount = computed(() => accounts.value.find(a => a.id === activeAccountId.value) || null)

  function addAccount(account) {
    accounts.value.push(account)
    activeAccountId.value = account.id
  }

  function removeAccount(id) {
    accounts.value = accounts.value.filter(a => a.id !== id)
    zones.value = zones.value.filter(z => z.accountId !== id)
    if (activeAccountId.value === id) {
      activeAccountId.value = accounts.value[0]?.id || null
    }
  }

  async function loadAllData() {
    try {
      const data = await window.go.main.App.LoadAllData()
      if (data.accounts) accounts.value = data.accounts
      if (data.zones) zones.value = data.zones
      if (data.certs) certList.value = data.certs
      if (accounts.value.length > 0 && !activeAccountId.value) {
        activeAccountId.value = accounts.value[0].id
      }
    } catch (e) {
      console.error('加载数据失败:', e)
    }
  }

  async function refreshZones(accountId) {
    try {
      const zoneData = await window.go.main.App.FetchZones(accountId)
      // Update zones for this account
      zones.value = zones.value.filter(z => z.accountId !== accountId)
      if (zoneData) {
        zones.value.push(...zoneData)
      }
    } catch (e) {
      console.error('获取域名列表失败:', e)
      throw e
    }
  }

  async function refreshCerts() {
    try {
      const certs = await window.go.main.App.ListCerts()
      certList.value = certs || []
    } catch (e) {
      console.error('获取证书列表失败:', e)
    }
  }

  return {
    accounts, activeAccountId, connected, activeAccount,
    addAccount, removeAccount,
    zones, certList,
    loadAllData, refreshZones, refreshCerts,
  }
})
