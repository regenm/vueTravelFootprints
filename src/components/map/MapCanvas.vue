<template>
  <div class="map-wrap">
    <div id="map-container" :class="{ picking: ui.addMode }"></div>

    <div v-if="loading || markers.loading" class="overlay">正在打开地图…</div>
    <div v-if="error" class="overlay error">
      <p>{{ error }}</p>
      <el-button type="primary" @click="initMap">重试</el-button>
    </div>

    <div v-if="ui.addMode" class="hint">
      点击地图上的位置来记录足迹
      <el-button size="small" @click="ui.addMode = false">取消</el-button>
    </div>

    <div v-if="showEmpty" class="empty-map">
      <h3>地图还是空的</h3>
      <p>记录第一处足迹，让旅途从这里开始。</p>
      <el-button type="primary" @click="ui.openForm()">开始记录</el-button>
    </div>
    <div v-else-if="showNoMatch" class="empty-map faint">
      <h3>没有符合条件的足迹</h3>
      <p>试试清空搜索或换一个分类。</p>
    </div>

    <div v-if="shareChip" class="share-chip">
      <div class="faces">
        <UserAvatar
          v-for="p in participants.slice(0, 5)"
          :key="p.id"
          :src="p.avatar"
          :name="p.displayName"
          :size="26"
          ring
        />
      </div>
      <div class="chip-text">
        <strong>{{ shareChip.title }}</strong>
        <small>{{ shareChip.hint }}</small>
      </div>
    </div>

    <div class="fab">
      <el-button circle @click="locateMe" title="我的位置">◎</el-button>
      <el-button circle @click="fitAll" title="查看全部足迹">⊕</el-button>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { createApp } from 'vue'
import { ElMessage } from 'element-plus'
import MarkerAvatar from './MarkerAvatar.vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { loadAMap } from '@/utils/amap'
import { useMarkersStore } from '@/stores/markers'
import { useUiStore } from '@/stores/ui'

const markers = useMarkersStore()
const ui = useUiStore()
const loading = ref(false)
const error = ref('')
let map = null
let markerApps = []
let amapMarkers = []

const showEmpty = computed(() => !loading.value && !markers.loading && !error.value && !markers.markers.length && !ui.addMode)
const showNoMatch = computed(() => !showEmpty.value && !markers.loading && markers.markers.length && !markers.filteredMarkers.length)

const participants = computed(() => markers.currentShare?.share?.participants || [])
const shareChip = computed(() => {
  const share = markers.currentShare?.share
  if (!share) return null
  const n = participants.value.length || 1
  const perm = markers.currentShare?.canEdit ? '可一起记录' : '只读'
  return {
    title: share.title,
    hint: `${n} 位旅人 · ${share.markerCount ?? markers.markers.length} 个地点 · ${perm}`
  }
})

async function initMap() {
  loading.value = true
  error.value = ''
  try {
    const AMap = await loadAMap()
    if (map) {
      map.destroy()
      map = null
    }
    map = new AMap.Map('map-container', {
      zoom: 5,
      center: [104.0, 35.0],
      mapStyle: 'amap://styles/fresh',
      viewMode: '2D'
    })
    map.addControl(new AMap.Scale())
    map.on('click', onMapClick)
    renderMarkers()
    fitAll()
  } catch (err) {
    error.value = `地图加载失败：${err.message}`
  } finally {
    loading.value = false
  }
}

function clearMarkers() {
  markerApps.forEach((app) => app.unmount())
  markerApps = []
  amapMarkers.forEach((m) => m.setMap(null))
  amapMarkers = []
}

function renderMarkers() {
  if (!map || !window.AMap) return
  clearMarkers()
  markers.filteredMarkers.forEach((item) => {
    const lng = Number(item.longitude)
    const lat = Number(item.latitude)
    if (Number.isNaN(lng) || Number.isNaN(lat)) return

    const el = document.createElement('div')
    const app = createApp(MarkerAvatar, {
      photo: item.photos?.[0] || '',
      category: item.category || '',
      active: item.id === markers.selectedId,
      showAuthor: markers.isShareView,
      authorAvatar: item.author?.avatar || '',
      authorName: item.author?.displayName || ''
    })
    app.mount(el)
    markerApps.push(app)

    const pin = new window.AMap.Marker({
      position: [lng, lat],
      content: el,
      offset: new window.AMap.Pixel(-21, -36),
      extData: item
    })
    pin.on('click', () => {
      markers.selectMarker(item)
    })
    pin.setMap(map)
    amapMarkers.push(pin)
  })
}

function onMapClick(e) {
  if (!ui.addMode) return
  const lng = e.lnglat.getLng()
  const lat = e.lnglat.getLat()
  const coords = { lng: lng.toFixed(6), lat: lat.toFixed(6), address: '', name: '' }

  if (window.AMap?.Geocoder) {
    const geocoder = new window.AMap.Geocoder()
    geocoder.getAddress([lng, lat], (status, result) => {
      if (status === 'complete' && result.regeocode) {
        coords.address = result.regeocode.formattedAddress || ''
        const poi = result.regeocode.pois?.[0]
        if (poi?.name) coords.name = poi.name
      }
      ui.openForm(null, coords)
    })
    return
  }
  ui.openForm(null, coords)
}

function fitAll() {
  if (!map || !amapMarkers.length) return
  map.setFitView(amapMarkers, false, [80, 80, 80, 80], 16)
}

function panTo(marker) {
  if (!map || !marker) return
  map.setZoomAndCenter(13, [Number(marker.longitude), Number(marker.latitude)])
}

function locateMe() {
  if (!navigator.geolocation) {
    ElMessage.warning('浏览器不支持定位')
    return
  }
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      const lng = pos.coords.longitude
      const lat = pos.coords.latitude
      map?.setZoomAndCenter(14, [lng, lat])
      if (ui.addMode) {
        ui.openForm(null, { lng: lng.toFixed(6), lat: lat.toFixed(6) })
      }
    },
    () => ElMessage.error('定位失败，请检查权限')
  )
}

watch(
  () => markers.filteredMarkers,
  () => {
    renderMarkers()
  },
  { deep: true }
)

watch(
  () => markers.selectedId,
  (id) => {
    renderMarkers()
    const item = markers.markers.find((m) => m.id === id)
    if (item) panTo(item)
  }
)

watch(
  () => ui.focusCoords,
  (coords) => {
    if (!coords || !map) return
    map.setZoomAndCenter(14, [Number(coords.lng), Number(coords.lat)])
  }
)

onMounted(initMap)
onBeforeUnmount(() => {
  clearMarkers()
  map?.destroy()
  map = null
})

defineExpose({ fitAll, renderMarkers })
</script>

<style scoped>
.map-wrap,
#map-container {
  width: 100%;
  height: 100%;
  position: relative;
}

#map-container.picking {
  cursor: crosshair;
}

.overlay,
.empty-map {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgba(244, 239, 230, 0.72);
  z-index: 5;
  text-align: center;
  padding: 24px;
}

.empty-map {
  align-content: center;
  gap: 10px;
}

.empty-map.faint {
  background: rgba(244, 239, 230, 0.62);
  pointer-events: none;
}

.empty-map h3 {
  font-family: var(--font-serif);
  font-size: 28px;
}

.hint {
  position: absolute;
  left: 50%;
  bottom: 28px;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  background: rgba(28, 25, 21, 0.82);
  color: #fff;
  border-radius: 999px;
  z-index: 6;
}

.share-chip {
  position: absolute;
  left: 16px;
  top: 16px;
  z-index: 6;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px 8px 10px;
  background: rgba(255, 253, 248, 0.94);
  border: 1px solid var(--tf-line);
  border-radius: 999px;
  box-shadow: var(--tf-shadow-soft);
  max-width: min(420px, calc(100% - 88px));
}

.faces {
  display: flex;
}

.faces :deep(.ua) {
  margin-left: -8px;
}

.faces :deep(.ua:first-child) {
  margin-left: 0;
}

.chip-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
  line-height: 1.2;
}

.chip-text strong {
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chip-text small {
  color: var(--tf-ink-faint);
  font-size: 11px;
}

.fab {
  position: absolute;
  right: 16px;
  bottom: 24px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  z-index: 6;
}

@media (max-width: 860px) {
  .share-chip {
    left: 10px;
    top: 10px;
    padding: 6px 12px 6px 8px;
  }
  .chip-text small { display: none; }
}
</style>
