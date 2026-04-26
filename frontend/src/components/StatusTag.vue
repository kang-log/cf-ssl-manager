<script setup>
const props = defineProps({
  status: { type: String, required: true }
})

const config = {
  valid:    { type: 'success', text: '有效' },
  expiring: { type: 'warning', text: '即将过期' },
  expired:  { type: 'danger',  text: '已过期' },
  pending:  { type: 'info',    text: '申请中' },
}

function getInfo() {
  return config[props.status] || { type: 'info', text: props.status }
}
</script>

<template>
  <el-tag :type="getInfo().type" size="small" effect="light">{{ getInfo().text }}</el-tag>
</template>
