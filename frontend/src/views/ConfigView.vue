<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '../stores/app'
import { ElMessage } from 'element-plus'

const store = useAppStore()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const form = ref({ email: '', apiKey: '', apiToken: '' })
const editForm = ref({ id: '', email: '', apiKey: '', apiToken: '' })
const showKey = ref(false)
const showEditKey = ref(false)
const testing = ref(false)
const editTesting = ref(false)
const filterAccountId = ref('')
const verifyingId = ref(null)
const zonesLoading = ref(false)
const debugInfo = ref(null)
const showDebugDialog = ref(false)

function openAddDialog() {
  form.value = { email: '', apiKey: '', apiToken: '' }
  showKey.value = false
  showAddDialog.value = true
}

function openEditDialog(acc) {
  editForm.value = { id: acc.id, email: acc.email, apiKey: '', apiToken: '' }
  showEditKey.value = false
  showEditDialog.value = true
}

async function testConnection() {
  if (!form.value.email || !form.value.apiKey) {
    return ElMessage.warning('请填写邮箱和 API Key')
  }
  testing.value = true
  try {
    const result = await window.go.main.App.AddAccount(
      form.value.email,
      form.value.apiKey,
      form.value.apiToken || ''
    )
    ElMessage.success(`账户 ${result.email} 添加成功！[debug: ${result.debugInfo || 'none'}]`)
    showAddDialog.value = false
    // Reload from backend to get full account data including hasApiKey/hasApiToken
    await store.loadAllData()
    store.activeAccountId = result.id
    await fetchZonesForAccount(result.id)
  } catch (e) {
    ElMessage.error('验证失败: ' + (e.message || e))
  } finally {
    testing.value = false
  }
}

async function submitEditAccount() {
  if (!editForm.value.apiKey) {
    return ElMessage.warning('请填写 API Key')
  }
  editTesting.value = true
  try {
    const result = await window.go.main.App.UpdateAccount(
      editForm.value.id,
      editForm.value.apiKey,
      editForm.value.apiToken || ''
    )
    ElMessage.success(`账户 ${result.email} 凭证已更新`)
    showEditDialog.value = false
    await store.loadAllData()
  } catch (e) {
    ElMessage.error('更新失败: ' + (e.message || e))
  } finally {
    editTesting.value = false
  }
}

async function fetchZonesForAccount(accountId) {
  zonesLoading.value = true
  try {
    await store.refreshZones(accountId)
    ElMessage.success('域名列表已更新')
  } catch (e) {
    ElMessage.error('获取域名列表失败: ' + (e.message || e))
  } finally {
    zonesLoading.value = false
  }
}

async function handleRemoveAccount(acc) {
  try {
    await window.go.main.App.RemoveAccount(acc.id)
    store.removeAccount(acc.id)
    ElMessage.success(`已移除账户 ${acc.email}`)
  } catch (e) {
    ElMessage.error('移除失败: ' + (e.message || e))
  }
}

async function verifyAccount(acc) {
  verifyingId.value = acc.id
  try {
    const ok = await window.go.main.App.VerifyAccount(acc.id)
    if (ok) {
      ElMessage.success(`账户 ${acc.email} 验证连接正常`)
    } else {
      ElMessage.error(`账户 ${acc.email} 验证失败`)
    }
  } catch (e) {
    ElMessage.error('验证失败: ' + (e.message || e))
  } finally {
    verifyingId.value = null
  }
}

function getAccountEmail(id) {
  return store.accounts.find(a => a.id === id)?.email || '未知'
}

function getZoneCount(accountId) {
  return store.zones.filter(z => z.accountId === accountId).length
}

const filteredZones = computed(() => {
  if (!filterAccountId.value) return store.zones
  return store.zones.filter(z => z.accountId === filterAccountId.value)
})

async function runDebug() {
  try {
    debugInfo.value = await window.go.main.App.DebugDB()
    showDebugDialog.value = true
  } catch (e) {
    ElMessage.error('调试失败: ' + (e.message || e))
  }
}

onMounted(() => {
  store.loadAllData()
})
</script>

<template>
  <div class="page">
    <div class="page-head">
      <div>
        <h2 class="page-title">Cloudflare 配置</h2>
        <p class="page-desc">管理 Cloudflare 账户凭证，程序将自动获取各账户下的域名列表</p>
      </div>
      <div style="display:flex;gap:8px">
        <el-button @click="runDebug" size="small" type="info" plain>诊断数据库</el-button>
        <el-button type="primary" @click="openAddDialog">
          <el-icon><Plus /></el-icon> 添加账户
        </el-button>
      </div>
    </div>

    <!-- 账户表格 -->
    <el-card shadow="never" class="section-card">
      <template #header>
        <div class="card-header">
          <span>账户列表</span>
          <el-tag type="info" size="small">{{ store.accounts.length }} 个</el-tag>
        </div>
      </template>
      <el-table :data="store.accounts" stripe style="width:100%" empty-text="暂未添加任何账户">
        <el-table-column label="邮箱" min-width="220">
          <template #default="{ row }">
            <div class="acc-email-cell">
              <el-icon :size="16" color="#409eff"><User /></el-icon>
              <span>{{ row.email }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="API Key" min-width="140">
          <template #default="{ row }">
            <el-tag v-if="row.hasApiKey" type="success" size="small" effect="plain">已配置</el-tag>
            <el-tag v-else type="info" size="small" effect="plain">未配置</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="API Token" min-width="140">
          <template #default="{ row }">
            <el-tag v-if="row.hasApiToken" type="success" size="small" effect="plain">已配置</el-tag>
            <el-tag v-else type="info" size="small" effect="plain">未配置</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="域名数" width="90" align="center">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ getZoneCount(row.id) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.id === store.activeAccountId ? 'success' : 'info'" size="small" effect="light">
              {{ row.id === store.activeAccountId ? '当前' : '已连接' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" align="center">
          <template #default="{ row }">
            <div class="action-btns">
              <el-button size="small" type="primary" plain @click="openEditDialog(row)">编辑</el-button>
              <el-button size="small" type="success" plain :loading="verifyingId === row.id" @click="verifyAccount(row)">
                {{ verifyingId === row.id ? '' : '验证' }}
              </el-button>
              <el-button size="small" type="warning" plain @click="store.activeAccountId = row.id" :disabled="row.id === store.activeAccountId">切换</el-button>
              <el-button size="small" type="danger" plain @click="handleRemoveAccount(row)">移除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 域名表格 -->
    <el-card shadow="never" class="section-card">
      <template #header>
        <div class="card-header">
          <span>已托管域名</span>
          <el-tag type="info" size="small">{{ filteredZones.length }} 个</el-tag>
        </div>
      </template>
      <div class="zone-filter">
        <span class="filter-label">按账户筛选：</span>
        <el-select v-model="filterAccountId" placeholder="全部账户" clearable style="width:260px" size="small">
          <el-option v-for="acc in store.accounts" :key="acc.id" :label="acc.email" :value="acc.id" />
        </el-select>
        <el-button size="small" type="primary" plain :loading="zonesLoading" @click="store.activeAccountId && fetchZonesForAccount(store.activeAccountId)">
          刷新域名列表
        </el-button>
      </div>
      <el-table :data="filteredZones" stripe size="small" style="width:100%">
        <el-table-column prop="name" label="域名" min-width="180" />
        <el-table-column label="所属账户" min-width="200">
          <template #default="{ row }">
            <span class="sub-text">{{ getAccountEmail(row.accountId) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'warning'" size="small" effect="plain">
              {{ row.status === 'active' ? '活跃' : row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="plan" label="Plan" width="90" />
        <el-table-column label="Nameserver" min-width="200">
          <template #default="{ row }">
            <span class="sub-text">{{ Array.isArray(row.ns) ? row.ns.join(', ') : row.ns }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加账户弹窗 -->
    <el-dialog v-model="showAddDialog" title="添加 Cloudflare 账户" width="480" destroy-on-close>
      <el-form :model="form" label-position="top">
        <el-form-item label="邮箱（Email）">
          <el-input v-model="form.email" placeholder="your@email.com" prefix-icon="Message" />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="form.apiKey" :type="showKey ? 'text' : 'password'" placeholder="Global API Key" prefix-icon="Key">
            <template #append>
              <el-button @click="showKey = !showKey">
                <el-icon><View v-if="!showKey" /><Hide v-else /></el-icon>
              </el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="API Token（可选，推荐）">
          <el-input v-model="form.apiToken" type="password" placeholder="Scoped API Token" prefix-icon="Key" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" :loading="testing" @click="testConnection" :disabled="!form.email || !form.apiKey">
          {{ testing ? '验证中...' : '验证并添加' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑账户弹窗 -->
    <el-dialog v-model="showEditDialog" title="编辑账户凭证" width="480" destroy-on-close>
      <el-form :model="editForm" label-position="top">
        <el-form-item label="邮箱">
          <el-input :value="editForm.email" disabled />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="editForm.apiKey" :type="showEditKey ? 'text' : 'password'" placeholder="填写新的 Global API Key" prefix-icon="Key">
            <template #append>
              <el-button @click="showEditKey = !showEditKey">
                <el-icon><View v-if="!showEditKey" /><Hide v-else /></el-icon>
              </el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="API Token（可选）">
          <el-input v-model="editForm.apiToken" type="password" placeholder="填写新的 Scoped API Token" prefix-icon="Key" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" :loading="editTesting" @click="submitEditAccount" :disabled="!editForm.apiKey">
          {{ editTesting ? '验证中...' : '验证并保存' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 调试弹窗 -->
    <el-dialog v-model="showDebugDialog" title="数据库诊断" width="560">
      <pre style="background:#1e1e1e;color:#d4d4d4;padding:16px;border-radius:6px;font-size:12px;line-height:1.6;max-height:400px;overflow:auto">{{ JSON.stringify(debugInfo, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<style scoped>
.page { max-width: 960px; }
.page-head { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; margin-bottom: 4px; }
.page-desc { font-size: 13px; color: #909399; }
.section-card { margin-bottom: 16px; }
.card-header { display: flex; align-items: center; gap: 8px; font-weight: 600; }
.acc-email-cell { display: flex; align-items: center; gap: 8px; font-weight: 500; }
.action-btns { display: flex; gap: 4px; justify-content: center; flex-wrap: wrap; }
.action-btns .el-button { margin-left: 0; }
.zone-filter { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.filter-label { font-size: 13px; color: #606266; flex-shrink: 0; }
.sub-text { font-size: 12px; color: #909399; }
</style>
