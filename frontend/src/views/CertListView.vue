<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/app'
import StatusTag from '../components/StatusTag.vue'
import { ElMessageBox, ElMessage } from 'element-plus'

const store = useAppStore()
const router = useRouter()
const search = ref('')
const filterCA = ref('')
const filterStatus = ref('')
const filterAccountId = ref('')
const sortField = ref('expiresAt')
const sortOrder = ref('asc')
const loading = ref(false)

function getRemainingDays(expiresAt) {
  const diff = new Date(expiresAt) - new Date()
  return Math.ceil(diff / (1000 * 60 * 60 * 24))
}

function remainingClass(days) {
  if (days <= 0) return 'days-expired'
  if (days <= 30) return 'days-warning'
  return 'days-ok'
}

function getAccountEmail(id) {
  return store.accounts.find(a => a.id === id)?.email || '未知'
}

const filteredList = computed(() => {
  let list = [...store.certList]
  if (filterAccountId.value) list = list.filter(c => c.accountId === filterAccountId.value)
  if (search.value) {
    const kw = search.value.toLowerCase()
    list = list.filter(c => c.domain.toLowerCase().includes(kw))
  }
  if (filterCA.value) list = list.filter(c => c.ca === filterCA.value)
  if (filterStatus.value) list = list.filter(c => c.status === filterStatus.value)
  list.sort((a, b) => {
    const va = a[sortField.value], vb = b[sortField.value]
    const cmp = va < vb ? -1 : va > vb ? 1 : 0
    return sortOrder.value === 'asc' ? cmp : -cmp
  })
  return list
})

const stats = computed(() => {
  const list = filteredList.value
  return {
    all: list.length,
    valid: list.filter(c => c.status === 'valid').length,
    expiring: list.filter(c => c.status === 'expiring').length,
    expired: list.filter(c => c.status === 'expired').length,
  }
})

function viewDetail(id) { router.push(`/certs/${id}`) }

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确定删除证书记录「${row.domain}」？`, '删除确认', { type: 'warning' })
    await window.go.main.App.DeleteCert(row.id)
    store.certList = store.certList.filter(c => c.id !== row.id)
    ElMessage.success('已删除')
  } catch (e) {
    if (e !== 'cancel' && e?.message !== 'cancel') {
      ElMessage.error('删除失败: ' + (e.message || e))
    }
  }
}

function copyDomain(domain) {
  navigator.clipboard?.writeText(domain)
  ElMessage.success('已复制')
}

onMounted(async () => {
  loading.value = true
  try {
    await store.refreshCerts()
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="page">
    <h2 class="page-title">证书列表</h2>

    <div class="stats-bar">
      <el-card shadow="never" class="stat-card">
        <div class="stat-num">{{ stats.all }}</div>
        <div class="stat-label">证书总数</div>
      </el-card>
      <el-card shadow="never" class="stat-card stat-valid">
        <div class="stat-num">{{ stats.valid }}</div>
        <div class="stat-label">有效</div>
      </el-card>
      <el-card shadow="never" class="stat-card stat-expiring">
        <div class="stat-num">{{ stats.expiring }}</div>
        <div class="stat-label">即将过期</div>
      </el-card>
      <el-card shadow="never" class="stat-card stat-expired">
        <div class="stat-num">{{ stats.expired }}</div>
        <div class="stat-label">已过期</div>
      </el-card>
    </div>

    <div class="filter-bar">
      <el-select v-model="filterAccountId" placeholder="全部账户" clearable style="width:200px">
        <el-option v-for="acc in store.accounts" :key="acc.id" :label="acc.email" :value="acc.id" />
      </el-select>
      <el-input v-model="search" placeholder="搜索域名..." prefix-icon="Search" clearable style="width:200px" />
      <el-select v-model="filterCA" placeholder="全部品牌" clearable style="width:140px">
        <el-option label="Let's Encrypt" value="Let's Encrypt" />
        <el-option label="LiteSSL" value="LiteSSL" />
      </el-select>
      <el-select v-model="filterStatus" placeholder="全部状态" clearable style="width:120px">
        <el-option label="有效" value="valid" />
        <el-option label="即将过期" value="expiring" />
        <el-option label="已过期" value="expired" />
      </el-select>
      <el-select v-model="sortField" style="width:140px">
        <el-option label="按到期时间" value="expiresAt" />
        <el-option label="按申请时间" value="issuedAt" />
        <el-option label="按域名" value="domain" />
      </el-select>
      <el-button-group>
        <el-button :type="sortOrder==='asc'?'primary':'info'" size="small" @click="sortOrder='asc'">↑</el-button>
        <el-button :type="sortOrder==='desc'?'primary':'info'" size="small" @click="sortOrder='desc'">↓</el-button>
      </el-button-group>
    </div>

    <el-card shadow="never">
      <el-table :data="filteredList" stripe style="width:100%" @row-dblclick="viewDetail" v-loading="loading">
        <el-table-column prop="domain" label="域名" min-width="170">
          <template #default="{ row }">
            <div class="domain-cell">
              <span class="domain-name">{{ row.domain }}</span>
              <el-button size="small" text type="primary" @click.stop="copyDomain(row.domain)">
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="账户" min-width="160">
          <template #default="{ row }">
            <span class="account-label">{{ getAccountEmail(row.accountId) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="sanCount" label="SAN" width="55" align="center" />
        <el-table-column prop="ca" label="品牌" width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ row.ca }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="keyAlgo" label="密钥" width="110" />
        <el-table-column prop="issuedAt" label="申请时间" width="150" />
        <el-table-column prop="expiresAt" label="到期时间" width="150" />
        <el-table-column label="剩余" width="80" align="center">
          <template #default="{ row }">
            <span :class="remainingClass(getRemainingDays(row.expiresAt))">
              {{ getRemainingDays(row.expiresAt) > 0 ? getRemainingDays(row.expiresAt) + '天' : '已过期' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <StatusTag :status="row.status" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-button size="small" type="primary" plain @click.stop="viewDetail(row.id)">查看</el-button>
              <el-button size="small" type="warning" plain @click.stop="ElMessage.info('续期功能开发中')">续期</el-button>
              <el-button size="small" type="danger" plain @click.stop="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.page-title { font-size: 20px; font-weight: 600; color: #303133; margin-bottom: 16px; }
.stats-bar { display: flex; gap: 12px; margin-bottom: 16px; }
.stat-card { flex: 1; text-align: center; padding: 4px 0; }
.stat-num { font-size: 28px; font-weight: 700; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }
.stat-valid .stat-num { color: #67c23a; }
.stat-expiring .stat-num { color: #e6a23c; }
.stat-expired .stat-num { color: #f56c6c; }
.filter-bar { display: flex; gap: 10px; margin-bottom: 16px; align-items: center; flex-wrap: wrap; }
.domain-cell { display: flex; align-items: center; gap: 4px; }
.domain-name { font-weight: 500; }
.account-label { font-size: 12px; color: #606266; }
.days-ok { color: #67c23a; font-weight: 600; }
.days-warning { color: #e6a23c; font-weight: 600; }
.days-expired { color: #f56c6c; font-weight: 600; }
.action-btns { display: flex; gap: 4px; justify-content: center; }
.action-btns .el-button { margin-left: 0; }
</style>
