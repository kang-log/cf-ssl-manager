<script setup>
import { ref, nextTick } from 'vue'

const props = defineProps({
  logs: { type: Array, default: () => [] }
})

const logRef = ref(null)

function scrollToBottom() {
  nextTick(() => {
    if (logRef.value) logRef.value.scrollTop = logRef.value.scrollHeight
  })
}

defineExpose({ scrollToBottom })
</script>

<template>
  <div class="log-panel" ref="logRef">
    <div v-for="(log, i) in logs" :key="i" class="log-line">
      <span class="log-time">{{ log.time }}</span>
      <el-tag :type="log.level === 'error' ? 'danger' : log.level === 'warn' ? 'warning' : 'info'" size="small" effect="plain" class="log-tag">
        {{ log.level === 'error' ? '错误' : log.level === 'warn' ? '警告' : '信息' }}
      </el-tag>
      <span class="log-msg">{{ log.msg }}</span>
    </div>
    <div v-if="logs.length === 0" class="log-empty">暂无日志</div>
  </div>
</template>

<style scoped>
.log-panel {
  background: #1e1e1e;
  color: #d4d4d4;
  border-radius: 6px;
  padding: 12px;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.8;
  height: 240px;
  overflow-y: auto;
}
.log-line { display: flex; align-items: center; gap: 8px; }
.log-time { color: #6a9955; flex-shrink: 0; }
.log-tag { transform: scale(0.85); }
.log-msg { word-break: break-all; }
.log-empty { color: #6a9955; text-align: center; padding: 40px 0; }
</style>
