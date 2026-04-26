<script setup>
import Sidebar from './components/Sidebar.vue'
</script>

<template>
  <div class="app-layout">
    <Sidebar />
    <div class="app-main">
      <router-view />
    </div>
  </div>
</template>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
html, body, #app { height: 100%; font-family: 'Microsoft YaHei', 'Segoe UI', sans-serif; }
.app-layout { display: flex; height: 100vh; background: #f0f2f5; }
.app-main { flex: 1; overflow-y: auto; padding: 24px; }
</style>
