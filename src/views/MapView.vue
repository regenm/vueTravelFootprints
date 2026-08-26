<template>
  <div class="app-shell">
    <AppHeader />
    <div class="app-main">
      <AppSidebar />
      <div class="map-stage">
        <MapCanvas />
      </div>
      <MarkerDetail v-if="ui.detailOpen && markers.selectedMarker" />
    </div>
    <MarkerForm />
    <ShareDialog />
    <ProfileDialog />
    <UserManager />
  </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import MapCanvas from '@/components/map/MapCanvas.vue'
import MarkerDetail from '@/components/marker/MarkerDetail.vue'
import MarkerForm from '@/components/marker/MarkerForm.vue'
import ShareDialog from '@/components/share/ShareDialog.vue'
import ProfileDialog from '@/components/layout/ProfileDialog.vue'
import UserManager from '@/components/admin/UserManager.vue'
import { useAuthStore } from '@/stores/auth'
import { useMarkersStore } from '@/stores/markers'
import { useUiStore } from '@/stores/ui'

const route = useRoute()
const auth = useAuthStore()
const markers = useMarkersStore()
const ui = useUiStore()

async function boot() {
  try {
    if (route.params.token) {
      await markers.loadShare(route.params.token)
      ui.sidebarOpen = window.innerWidth > 860
      return
    }
    if (auth.isLoggedIn) {
      await markers.fetchMarkers()
    }
  } catch (err) {
    ElMessage.error(err.message || '加载足迹失败')
  }
}

function onKey(e) {
  if (e.key === 'Escape') {
    ui.addMode = false
    ui.detailOpen = false
  }
}

watch(() => route.fullPath, boot)
onMounted(() => {
  boot()
  window.addEventListener('keydown', onKey)
})
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
</script>
