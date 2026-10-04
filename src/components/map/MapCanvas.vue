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
      <el-button
        circle
        :type="ui.regionLayerVisible ? 'primary' : 'default'"
        @click="ui.regionLayerVisible = !ui.regionLayerVisible"
        title="城市填色开关"
      >◉</el-button>
      <el-button
        circle
        :type="satelliteOn ? 'primary' : 'default'"
        @click="toggleSatellite"
        title="卫星图层开关"
      >▦</el-button>
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
import { wgs84ToGcj02 } from '@/utils/coords'
import { useMarkersStore } from '@/stores/markers'
import { useUiStore } from '@/stores/ui'

// 库中 WGS-84 → 高德 GCJ-02，用于地图视角与标记定位
function toAmapPos(lng, lat) {
  const [gLng, gLat] = wgs84ToGcj02(Number(lng), Number(lat))
  return [gLng, gLat]
}

const markers = useMarkersStore()
const ui = useUiStore()
const loading = ref(false)
const error = ref('')
let map = null
let markerApps = []
let amapMarkers = []
let regionLayer = null
// 卫星影像 + 路网标注两层一起切，纯卫星图没有道路和地名文字
const satelliteOn = ref(false)
let satelliteLayers = []

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
      mapStyle: 'amap://styles/light',
      viewMode: '2D'
    })
    map.addControl(new AMap.Scale())
    map.on('click', onMapClick)
    if (window.AMap.TileLayer?.Satellite) {
      satelliteLayers = [new window.AMap.TileLayer.Satellite(), new window.AMap.TileLayer.RoadNet()]
      if (satelliteOn.value) satelliteLayers.forEach((l) => l.setMap(map))
    }
    if (window.AMap.DistrictLayer?.Province) {
      regionLayer = createRegionLayer()
      syncRegionLayer()
    }
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
    const lngNum = Number(item.longitude)
    const latNum = Number(item.latitude)
    if (Number.isNaN(lngNum) || Number.isNaN(latNum)) return
    // 库中 WGS-84 → 高德 GCJ-02 再渲染
    const [lng, lat] = wgs84ToGcj02(lngNum, latNum)

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
  map.setZoomAndCenter(13, toAmapPos(marker.longitude, marker.latitude))
}

function toggleSatellite() {
  satelliteOn.value = !satelliteOn.value
  satelliteLayers.forEach((l) => l.setMap(satelliteOn.value ? map : null))
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
      // 浏览器定位是 WGS-84：地图视角需转 GCJ-02；表单约定 GCJ-02，同样传转换后的值，
      // 提交时由表单统一 gcj02ToWgs84 转回 WGS-84 入库（简报此行原文传 lng/lat，与其
      // “三个入口收敛到同一条转换路径”的核心约定冲突，按约定修正，详见任务报告）
      const [gLng, gLat] = wgs84ToGcj02(lng, lat)
      map?.setZoomAndCenter(14, [gLng, gLat])
      if (ui.addMode) {
        ui.openForm(null, { lng: gLng.toFixed(6), lat: gLat.toFixed(6) })
      }
    },
    () => ElMessage.error('定位失败，请检查权限')
  )
}

watch(
  () => markers.filteredMarkers,
  () => {
    renderMarkers()
    syncRegionLayer()
  },
  { deep: true }
)

watch(
  () => ui.regionLayerVisible,
  () => {
    syncRegionLayer()
  }
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
    map.setZoomAndCenter(14, toAmapPos(coords.lng, coords.lat))
  }
)

onMounted(initMap)
onBeforeUnmount(() => {
  clearMarkers()
  map?.destroy()
  map = null
})

// 创建区县填色图层：DistrictLayer.Province 在 depth 2 才下发区县级要素
// （实测 Country 图层与 Province depth 3 均只给到市级，区县码永不命中）；
// adcode 限定为到访省份，fill 函数式样式按 props.adcode 匹配已到访区县，
// 未到访区县返回完全透明，只点亮去过的区县
function createRegionLayer() {
  if (!map) return null
  const provinces = markers.visitedProvinceCodes.map(String)
  if (!provinces.length) return null
  const visited = new Set(markers.visitedDistrictCodes.map(String))
  const layer = new window.AMap.DistrictLayer.Province({
    zIndex: 99,
    adcode: provinces,
    depth: 2,
    styles: {
      fill: (props) => (visited.has(String(props?.adcode)) ? 'rgba(72, 119, 173, 0.55)' : 'rgba(0, 0, 0, 0)')
    }
  })
  layer.setMap(map)
  return layer
}

// 已到访集合变化时函数式样式无法增量更新，销毁重建图层；同时同步显隐开关。
// JSAPI 未就绪时（足迹先于地图加载完成）跳过，initMap 完成后会自建图层
function syncRegionLayer() {
  if (!map || !window.AMap?.DistrictLayer?.Province) return
  regionLayer?.setMap(null)
  regionLayer?.destroy?.()
  regionLayer = createRegionLayer()
  regionLayer?.[ui.regionLayerVisible ? 'show' : 'hide']?.()
}

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
