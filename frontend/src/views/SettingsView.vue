<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const settings = ref({
  certPath: 'C:/Users/user/cf-ssl-certs',
  reminderDays: '30',
  autoRenew: 'false',
  logLevel: 'info',
  proxyEnabled: 'false',
  proxyType: 'http',
  proxyAddr: '',
})
const saving = ref(false)
const selectingDir = ref(false)

onMounted(async () => {
  try {
    const data = await window.go.main.App.GetSettings()
    if (data.certPath) settings.value.certPath = data.certPath
    if (data.reminderDays) settings.value.reminderDays = data.reminderDays
    if (data.autoRenew) settings.value.autoRenew = data.autoRenew
    if (data.logLevel) settings.value.logLevel = data.logLevel
    if (data.proxyEnabled) settings.value.proxyEnabled = data.proxyEnabled
    if (data.proxyType) settings.value.proxyType = data.proxyType
    if (data.proxyAddr) settings.value.proxyAddr = data.proxyAddr
  } catch (e) {
    console.error('加载设置失败:', e)
  }
})

async function selectDirectory() {
  selectingDir.value = true
  try {
    const dir = await window.go.main.App.SelectDirectory()
    if (dir) {
      settings.value.certPath = dir
    }
  } catch (e) {
    console.error('选择目录失败:', e)
  } finally {
    selectingDir.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    await window.go.main.App.SaveSettings(JSON.stringify(settings.value))
    ElMessage.success('设置已保存')
  } catch (e) {
    ElMessage.error('保存失败: ' + (e.message || e))
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="page">
    <h2 class="page-title">设置</h2>

    <el-card shadow="never" class="settings-card">
      <template #header><span class="card-title">证书存储</span></template>
      <el-form label-width="140px">
        <el-form-item label="默认存储路径">
          <el-input v-model="settings.certPath" placeholder="证书文件保存目录">
            <template #append>
              <el-button @click="selectDirectory" :loading="selectingDir">浏览</el-button>
            </template>
          </el-input>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="settings-card">
      <template #header><span class="card-title">到期提醒</span></template>
      <el-form label-width="140px">
        <el-form-item label="提前提醒天数">
          <el-input-number v-model="settings.reminderDays" :min="1" :max="90" />
          <span class="form-tip">天前开始提醒</span>
        </el-form-item>
        <el-form-item label="自动续期">
          <el-switch v-model="settings.autoRenew" />
          <span class="form-tip">到期前 30 天自动续期</span>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="settings-card">
      <template #header><span class="card-title">网络代理</span></template>
      <el-form label-width="140px">
        <el-form-item label="启用代理">
          <el-switch v-model="settings.proxyEnabled" />
        </el-form-item>
        <template v-if="settings.proxyEnabled">
          <el-form-item label="代理类型">
            <el-select v-model="settings.proxyType" style="width:160px">
              <el-option label="HTTP" value="http" />
              <el-option label="SOCKS5" value="socks5" />
            </el-select>
          </el-form-item>
          <el-form-item label="代理地址">
            <el-input v-model="settings.proxyAddr" placeholder="127.0.0.1:7890" />
          </el-form-item>
        </template>
      </el-form>
    </el-card>

    <el-card shadow="never" class="settings-card">
      <template #header><span class="card-title">日志</span></template>
      <el-form label-width="140px">
        <el-form-item label="日志级别">
          <el-select v-model="settings.logLevel" style="width:160px">
            <el-option label="Debug" value="debug" />
            <el-option label="Info" value="info" />
            <el-option label="Warn" value="warn" />
            <el-option label="Error" value="error" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="settings-card">
      <template #header><span class="card-title">关于</span></template>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="软件名称">CF SSL Manager</el-descriptions-item>
        <el-descriptions-item label="版本">v1.0.0</el-descriptions-item>
        <el-descriptions-item label="技术栈">Go + Wails v2 + Vue3</el-descriptions-item>
        <el-descriptions-item label="支持平台">Windows 10/11 (64-bit)</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <div style="margin-top:20px">
      <el-button type="primary" @click="saveSettings" :loading="saving">保存设置</el-button>
    </div>
  </div>
</template>

<style scoped>
.page { max-width: 680px; }
.page-title { font-size: 20px; font-weight: 600; color: #303133; margin-bottom: 16px; }
.settings-card { margin-bottom: 16px; }
.card-title { font-weight: 600; font-size: 14px; }
.form-tip { font-size: 12px; color: #909399; margin-left: 12px; }
</style>
