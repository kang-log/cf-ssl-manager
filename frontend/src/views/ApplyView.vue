<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '../stores/app'
import LogPanel from '../components/LogPanel.vue'
import { ElMessage } from 'element-plus'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const store = useAppStore()

const form = ref({
  accountId: '',
  zoneId: '',
  domains: '',
  ca: "Let's Encrypt",
  keyAlgo: 'ECDSA P-256',
  env: 'production',
  customCertPath: '',
})

const applying = ref(false)
const logs = ref([])
const logPanel = ref(null)
const selectDirLoading = ref(false)

const selectedZone = computed(() => store.zones.find(z => z.id === form.value.zoneId))

const accountZones = computed(() => {
  if (!form.value.accountId) return []
  return store.zones.filter(z => z.accountId === form.value.accountId)
})

function addLog(level, msg) {
  const now = new Date()
  const time = `${String(now.getHours()).padStart(2,'0')}:${String(now.getMinutes()).padStart(2,'0')}:${String(now.getSeconds()).padStart(2,'0')}`
  logs.value.push({ time, level, msg })
  logPanel.value?.scrollToBottom()
}

onMounted(() => {
  EventsOn('acme-log', (data) => {
    addLog(data.level || 'info', data.msg || '')
  })
})

async function selectCertDir() {
  selectDirLoading.value = true
  try {
    const dir = await window.go.main.App.SelectDirectory()
    if (dir) {
      form.value.customCertPath = dir
    }
  } catch (e) {
    console.error(e)
  } finally {
    selectDirLoading.value = false
  }
}

async function applyCert() {
  if (!form.value.accountId) return ElMessage.warning('请先添加并选择账户')
  if (!form.value.zoneId) return ElMessage.warning('请选择域名')
  if (!form.value.domains.trim()) return ElMessage.warning('请输入证书域名')

  applying.value = true
  logs.value = []

  try {
    const result = await window.go.main.App.ApplyCert(
      form.value.accountId,
      form.value.zoneId,
      form.value.domains.trim(),
      form.value.ca,
      form.value.keyAlgo,
      form.value.env,
      form.value.customCertPath || ''
    )
    addLog('info', '证书申请完成！')
    ElMessage.success('证书申请成功')
    store.refreshCerts()
  } catch (e) {
    addLog('error', '申请失败: ' + (e.message || e))
    ElMessage.error('证书申请失败: ' + (e.message || e))
  } finally {
    applying.value = false
  }
}
</script>

<template>
  <div class="page">
    <h2 class="page-title">申请证书</h2>
    <p class="page-desc">选择域名、填写需要申请证书的域名，程序将自动完成 DNS 验证和证书签发。</p>

    <div class="apply-layout">
      <el-card shadow="never" class="apply-form-card">
        <el-form :model="form" label-position="top">
          <el-form-item label="选择账户">
            <el-select v-model="form.accountId" placeholder="请选择账户" style="width:100%" filterable @change="form.zoneId = ''">
              <el-option v-for="acc in store.accounts" :key="acc.id" :label="acc.email" :value="acc.id" />
            </el-select>
          </el-form-item>

          <el-form-item label="选择托管域名">
            <el-select v-model="form.zoneId" placeholder="请选择域名" style="width:100%" filterable :disabled="!form.accountId">
              <el-option v-for="z in accountZones" :key="z.id" :label="z.name" :value="z.id">
                <span>{{ z.name }}</span>
                <el-tag size="small" type="info" style="margin-left:8px">{{ z.plan }}</el-tag>
              </el-option>
            </el-select>
          </el-form-item>

          <el-form-item label="证书域名（每行一个，支持泛域名）">
            <el-input v-model="form.domains" type="textarea" :rows="4" placeholder="example.com&#10;*.example.com&#10;www.example.com" />
            <div class="form-tip" v-if="selectedZone">
              提示：域名必须属于 <b>{{ selectedZone.name }}</b>，泛域名格式为 *.{{ selectedZone.name }}
            </div>
          </el-form-item>

          <el-row :gutter="16">
            <el-col :span="8">
              <el-form-item label="证书品牌">
                <el-select v-model="form.ca" style="width:100%">
                  <el-option label="Let's Encrypt" value="Let's Encrypt" />
                  <el-option label="LiteSSL" value="LiteSSL" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="密钥算法">
                <el-select v-model="form.keyAlgo" style="width:100%">
                  <el-option label="ECDSA P-256" value="ECDSA P-256" />
                  <el-option label="ECDSA P-384" value="ECDSA P-384" />
                  <el-option label="RSA 2048" value="RSA 2048" />
                  <el-option label="RSA 4096" value="RSA 4096" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="环境">
                <el-select v-model="form.env" style="width:100%">
                  <el-option label="Production（正式）" value="production" />
                  <el-option label="Staging（测试）" value="staging" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="证书保存路径（可选，留空使用默认路径）">
            <el-input v-model="form.customCertPath" placeholder="留空则使用设置中的默认路径">
              <template #append>
                <el-button @click="selectCertDir" :loading="selectDirLoading">浏览</el-button>
              </template>
            </el-input>
            <div class="form-tip">自定义本次证书的保存目录，不填写则使用设置页配置的默认路径</div>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" size="large" :loading="applying" @click="applyCert" :disabled="!store.connected" style="width:100%">
              <el-icon><DocumentAdd /></el-icon>
              {{ applying ? '申请中...' : '申请证书' }}
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card shadow="never" class="apply-log-card">
        <template #header>
          <div class="log-header">
            <span>申请日志</span>
            <el-button size="small" text type="info" @click="logs = []">清空</el-button>
          </div>
        </template>
        <LogPanel :logs="logs" ref="logPanel" />
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.page { max-width: 900px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; margin-bottom: 8px; }
.page-desc { font-size: 13px; color: #909399; margin-bottom: 24px; }
.apply-layout { display: flex; flex-direction: column; gap: 20px; }
.form-tip { font-size: 12px; color: #909399; margin-top: 4px; }
.log-header { display: flex; justify-content: space-between; align-items: center; font-weight: 600; }
</style>
