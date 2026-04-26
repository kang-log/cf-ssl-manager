import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/config' },
  { path: '/config', name: 'Config', component: () => import('../views/ConfigView.vue') },
  { path: '/apply', name: 'Apply', component: () => import('../views/ApplyView.vue') },
  { path: '/certs', name: 'Certs', component: () => import('../views/CertListView.vue') },
  { path: '/certs/:id', name: 'CertDetail', component: () => import('../views/CertDetailView.vue') },
  { path: '/settings', name: 'Settings', component: () => import('../views/SettingsView.vue') },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
