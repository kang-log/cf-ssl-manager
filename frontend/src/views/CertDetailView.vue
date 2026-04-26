<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '../stores/app'
import StatusTag from '../components/StatusTag.vue'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const store = useAppStore()
const pemTab = ref('cert')
const showKey = ref(false)
const remoteChecking = ref(false)
const remoteHost = ref('')
const remotePort = ref('443')
const loading = ref(false)
const certDetail = ref(null)

const cert = computed(() => store.certList.find(c => c.id === Number(route.params.id)))

function getRemainingDays(expiresAt) {
  const diff = new Date(expiresAt) - new Date()
  return Math.ceil(diff / (1000 * 60 * 60 * 24))
}

function copyText(text) {
  navigator.clipboard?.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

async function checkRemote() {
  if (!remoteHost.value) return ElMessage.warning('请输入域名')
  remoteChecking.value = true
  try {
    const result = await window.go.main.App.CheckRemoteCert(remoteHost.value, remotePort.value)
    ElMessage.success('远程证书检测完成')
    console.log('Remote cert:', result)
  } catch (e) {
    ElMessage.error('检测失败: ' + (e.message || e))
  } finally {
    remoteChecking.value = false
  }
}

onMounted(async () => {
  if (!cert.value) {
    router.push('/certs')
    return
  }
  // Load detailed cert info from backend
  loading.value = true
  try {
    certDetail.value = await window.go.main.App.GetCertDetail(cert.value.id)
  } catch (e) {
    console.error('获取证书详情失败:', e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="detail-page" v-if="cert" v-loading="loading">
    <!-- 顶部信息栏 -->
    <div class="detail-top">
      <div class="detail-top-left">
        <el-button text type="primary" @click="router.push('/certs')" class="back-btn">
          <el-icon><ArrowLeft /></el-icon> 返回列表
        </el-button>
        <div class="detail-title-row">
          <h2 class="detail-title">{{ cert.domain }}</h2>
          <StatusTag :status="cert.status" />
          <el-tag size="small" effect="plain">{{ cert.ca }}</el-tag>
        </div>
        <div class="detail-sub">
          申请时间 {{ cert.issuedAt }} · 到期时间 {{ cert.expiresAt }} · 剩余
          <span :class="getRemainingDays(cert.expiresAt) <= 30 ? 'days-warn' : 'days-ok'">
            {{ getRemainingDays(cert.expiresAt) > 0 ? getRemainingDays(cert.expiresAt) + ' 天' : '已过期' }}
          </span>
        </div>
      </div>
    </div>

    <!-- 主体两栏 -->
    <div class="detail-body">
      <!-- 左栏 -->
      <div class="detail-col">
        <el-card shadow="never" class="info-card">
          <template #header><div class="card-h"><el-icon><InfoFilled /></el-icon> 基本信息</div></template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="主题 (Subject)">
              <code class="mono">{{ certDetail?.detail?.subject || 'CN=' + cert.domain }}</code>
            </el-descriptions-item>
            <el-descriptions-item label="颁发者 (Issuer)">
              <code class="mono">{{ certDetail?.detail?.issuer || '加载中...' }}</code>
            </el-descriptions-item>
            <el-descriptions-item label="序列号">
              <code class="mono">{{ certDetail?.detail?.serialNumber || '加载中...' }}</code>
            </el-descriptions-item>
            <el-descriptions-item label="签名算法">{{ certDetail?.detail?.sigAlgorithm || '加载中...' }}</el-descriptions-item>
            <el-descriptions-item label="公钥算法">{{ certDetail?.detail?.keyAlgorithm || cert.keyAlgo }}</el-descriptions-item>
            <el-descriptions-item label="生效时间">{{ cert.issuedAt }}</el-descriptions-item>
            <el-descriptions-item label="到期时间">{{ cert.expiresAt }}</el-descriptions-item>
            <el-descriptions-item label="剩余天数">
              <span :class="getRemainingDays(cert.expiresAt) <= 30 ? 'days-warn' : 'days-ok'" style="font-weight:700">
                {{ getRemainingDays(cert.expiresAt) > 0 ? getRemainingDays(cert.expiresAt) + ' 天' : '已过期 ' + Math.abs(getRemainingDays(cert.expiresAt)) + ' 天' }}
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never" class="info-card">
          <template #header><div class="card-h"><el-icon><Key /></el-icon> 指纹信息</div></template>
          <div class="fp-item">
            <div class="fp-head">
              <span class="fp-label">SHA-1</span>
              <el-button size="small" text type="primary" @click="copyText(certDetail?.detail?.sha1 || '')">
                <el-icon><CopyDocument /></el-icon> 复制
              </el-button>
            </div>
            <code class="fp-value">{{ certDetail?.detail?.sha1 || '加载中...' }}</code>
          </div>
          <el-divider style="margin:12px 0" />
          <div class="fp-item">
            <div class="fp-head">
              <span class="fp-label">SHA-256</span>
              <el-button size="small" text type="primary" @click="copyText(certDetail?.detail?.sha256 || '')">
                <el-icon><CopyDocument /></el-icon> 复制
              </el-button>
            </div>
            <code class="fp-value">{{ certDetail?.detail?.sha256 || '加载中...' }}</code>
          </div>
        </el-card>

        <el-card shadow="never" class="info-card">
          <template #header><div class="card-h"><el-icon><List /></el-icon> SAN 域名列表</div></template>
          <div class="san-grid">
            <div v-for="(san, i) in (certDetail?.detail?.sanList || cert.sanList)" :key="i" class="san-chip">
              <el-icon color="#409eff" :size="14"><CircleCheck /></el-icon>
              <span>{{ san }}</span>
              <el-tag v-if="san.includes('*')" type="warning" size="small" effect="plain">泛域名</el-tag>
            </div>
          </div>
        </el-card>

        <el-card shadow="never" class="info-card">
          <template #header><div class="card-h"><el-icon><Share /></el-icon> 证书信任链</div></template>
          <div class="chain-tree">
            <div class="chain-node">
              <div class="chain-icon chain-ok">
                <el-icon :size="14"><Check /></el-icon>
              </div>
              <div>
                <div class="chain-name">根证书</div>
                <div class="chain-sub">{{ cert.ca === 'LiteSSL' ? 'LiteSSL Root CA' : 'ISRG Root X1 (Let\'s Encrypt)' }}</div>
              </div>
            </div>
            <div class="chain-connector"></div>
            <div class="chain-node">
              <div class="chain-icon chain-ok">
                <el-icon :size="14"><Check /></el-icon>
              </div>
              <div>
                <div class="chain-name">中间证书</div>
                <div class="chain-sub">{{ cert.ca }} 中间 CA</div>
              </div>
            </div>
            <div class="chain-connector"></div>
            <div class="chain-node">
              <div class="chain-icon chain-leaf">
                <el-icon :size="14"><Check /></el-icon>
              </div>
              <div>
                <div class="chain-name">{{ cert.domain }}</div>
                <div class="chain-sub">域名证书</div>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 右栏：证书内容 -->
      <div class="detail-col">
        <el-card shadow="never" class="pem-card">
          <template #header>
            <div class="card-h">
              <span><el-icon><Document /></el-icon> 证书 / 密钥内容</span>
              <el-button size="small" text type="primary" @click="copyText(pemTab === 'cert' ? (cert.certPem || '') : pemTab === 'chain' ? (cert.chainPem || '') : pemTab === 'fullchain' ? (cert.fullChainPem || '') : (cert.keyPem || ''))">
                <el-icon><CopyDocument /></el-icon> 复制
              </el-button>
            </div>
          </template>
          <el-tabs v-model="pemTab" type="border-card" class="pem-tabs">
            <el-tab-pane name="cert">
              <template #label>
                <span class="tab-label"><el-icon><Document /></el-icon> 域名证书</span>
              </template>
              <pre class="pem-block">{{ cert.certPem || '暂无数据' }}</pre>
            </el-tab-pane>
            <el-tab-pane name="chain">
              <template #label>
                <span class="tab-label"><el-icon><Connection /></el-icon> 中间证书</span>
              </template>
              <pre class="pem-block">{{ cert.chainPem || '暂无数据' }}</pre>
            </el-tab-pane>
            <el-tab-pane name="fullchain">
              <template #label>
                <span class="tab-label"><el-icon><Files /></el-icon> 完整证书链</span>
              </template>
              <pre class="pem-block">{{ cert.fullChainPem || '暂无数据' }}</pre>
            </el-tab-pane>
            <el-tab-pane name="key">
              <template #label>
                <span class="tab-label"><el-icon><Lock /></el-icon> 私钥 (KEY)</span>
              </template>
              <div v-if="!showKey" class="key-mask">
                <el-icon :size="40" color="#e6a23c"><Lock /></el-icon>
                <p class="key-mask-title">私钥内容已隐藏</p>
                <p class="key-mask-desc">私钥是敏感信息，请确认当前环境安全后再查看</p>
                <el-button type="warning" @click="showKey = true">
                  <el-icon><View /></el-icon> 显示私钥
                </el-button>
              </div>
              <div v-else>
                <div class="key-warning">
                  <el-icon color="#e6a23c"><WarningFilled /></el-icon>
                  <span>请勿泄露私钥，私钥泄露将导致证书安全性失效</span>
                </div>
                <pre class="pem-block pem-key">{{ cert.keyPem || '暂无数据' }}</pre>
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-card>

        <el-card shadow="never" class="info-card">
          <template #header><div class="card-h"><el-icon><Connection /></el-icon> 在线证书检测</div></template>
          <div class="remote-check">
            <el-input v-model="remoteHost" placeholder="输入域名" style="flex:1" />
            <el-input v-model="remotePort" placeholder="端口" style="width:80px" />
            <el-button type="primary" :loading="remoteChecking" @click="checkRemote">
              <el-icon><Connection /></el-icon> 检测
            </el-button>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.detail-page { max-width: 1100px; }

.detail-top { margin-bottom: 20px; }
.back-btn { padding: 0; margin-bottom: 8px; }
.detail-title-row { display: flex; align-items: center; gap: 10px; }
.detail-title { font-size: 22px; font-weight: 700; color: #1a1a2e; margin: 0; }
.detail-sub { font-size: 13px; color: #909399; margin-top: 6px; }
.days-ok { color: #67c23a; font-weight: 700; }
.days-warn { color: #e6a23c; font-weight: 700; }

.detail-body { display: flex; gap: 20px; align-items: flex-start; }
.detail-col { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 16px; }

.info-card :deep(.el-card__header) { padding: 12px 16px; background: #fafbfc; }
.card-h { display: flex; align-items: center; justify-content: space-between; font-weight: 600; font-size: 14px; color: #303133; gap: 6px; }
.card-h .el-icon { color: #409eff; }

.mono { font-family: Consolas, monospace; font-size: 12px; color: #606266; background: #f5f7fa; padding: 1px 6px; border-radius: 3px; word-break: break-all; }

.fp-item {}
.fp-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.fp-label { font-size: 12px; font-weight: 600; color: #606266; }
.fp-value { font-family: Consolas, monospace; font-size: 11px; color: #909399; word-break: break-all; line-height: 1.6; margin: 0; }

.san-grid { display: flex; flex-direction: column; gap: 8px; }
.san-chip { display: flex; align-items: center; gap: 6px; padding: 6px 10px; background: #f5f7fa; border-radius: 6px; font-size: 13px; font-weight: 500; }

.chain-tree { padding: 4px 0; }
.chain-node { display: flex; align-items: center; gap: 12px; padding: 6px 0; }
.chain-icon { width: 26px; height: 26px; border-radius: 50%; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.chain-ok { background: #f0f9eb; color: #67c23a; }
.chain-leaf { background: #ecf5ff; color: #409eff; }
.chain-name { font-weight: 600; font-size: 13px; color: #303133; }
.chain-sub { font-size: 12px; color: #909399; }
.chain-connector { border-left: 2px solid #dcdfe6; height: 14px; margin-left: 12px; }

.pem-card :deep(.el-card__header) { padding: 12px 16px; background: #fafbfc; }
.pem-tabs :deep(.el-tabs__content) { padding: 0; }
.pem-tabs :deep(.el-tabs__header) { margin: 0; }
.tab-label { display: flex; align-items: center; gap: 4px; }
.pem-block {
  background: #1b2838;
  color: #a8c7c7;
  border-radius: 0;
  padding: 16px;
  font-size: 12px;
  line-height: 1.7;
  overflow-x: auto;
  max-height: 420px;
  overflow-y: auto;
  font-family: 'Cascadia Code', 'Fira Code', Consolas, monospace;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
.pem-key { color: #e6a23c; }

.key-mask { display: flex; flex-direction: column; align-items: center; gap: 10px; padding: 48px 24px; }
.key-mask-title { font-size: 15px; font-weight: 600; color: #303133; margin: 0; }
.key-mask-desc { font-size: 12px; color: #909399; margin: 0; }

.key-warning { display: flex; align-items: center; gap: 8px; padding: 8px 12px; background: #fdf6ec; color: #e6a23c; font-size: 12px; border-bottom: 1px solid #faecd8; }

.remote-check { display: flex; gap: 10px; align-items: center; }
</style>
