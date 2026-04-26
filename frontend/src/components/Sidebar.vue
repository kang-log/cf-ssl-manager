<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '../stores/app'

const route = useRoute()
const router = useRouter()
const store = useAppStore()

const activeMenu = computed(() => {
  if (route.path.startsWith('/certs')) return '/certs'
  return route.path
})

function goTo(path) {
  router.push(path)
}
</script>

<template>
  <div class="sidebar">
    <div class="sidebar-header">
      <el-icon :size="28" color="#409eff"><Lock /></el-icon>
      <span class="sidebar-title">CF SSL Manager</span>
    </div>

    <div class="sidebar-status" v-if="store.connected">
      <el-tag type="success" size="small" effect="plain">{{ store.accounts.length }} 账户</el-tag>
      <span class="sidebar-email" v-if="store.activeAccount">{{ store.activeAccount.email }}</span>
    </div>
    <div class="sidebar-status" v-else>
      <el-tag type="info" size="small" effect="plain">未连接</el-tag>
    </div>

    <el-menu :default-active="activeMenu" class="sidebar-menu" @select="goTo">
      <el-menu-item index="/config">
        <el-icon><Connection /></el-icon>
        <span>Cloudflare 配置</span>
      </el-menu-item>
      <el-menu-item index="/apply">
        <el-icon><DocumentAdd /></el-icon>
        <span>申请证书</span>
      </el-menu-item>
      <el-menu-item index="/certs">
        <el-icon><Files /></el-icon>
        <span>证书列表</span>
      </el-menu-item>
      <el-menu-item index="/settings">
        <el-icon><Setting /></el-icon>
        <span>设置</span>
      </el-menu-item>
    </el-menu>

    <div class="sidebar-footer">
      <span>v1.0.0</span>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  width: 220px;
  background: #fff;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #e4e7ed;
  user-select: none;
}
.sidebar-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 20px 12px;
  font-size: 16px;
  font-weight: 700;
  color: #303133;
}
.sidebar-status {
  padding: 0 20px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.sidebar-email {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sidebar-menu {
  flex: 1;
  border-right: none;
}
.sidebar-menu .el-menu-item {
  height: 48px;
  font-size: 14px;
}
.sidebar-footer {
  padding: 16px 20px;
  text-align: center;
  font-size: 12px;
  color: #c0c4cc;
  border-top: 1px solid #ebeef5;
}
</style>
