<template>
  <div id="map-container" :class="{ 'add-mode-cursor': markersStore.addMode }">
    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner">
        <el-icon class="is-loading"><Loading /></el-icon>
        加载中...
      </div>
    </div>
    <div v-if="error" class="error-overlay">
      <div class="error-message">
        <p>{{ error }}</p>
        <el-button @click="initMap" type="primary">重试</el-button>
      </div>
    </div>
    <div v-if="markersStore.addMode" class="add-mode-hint-overlay">
      <div class="add-mode-hint-box">
        <el-icon><Plus /></el-icon>
        <span>右键地图添加足迹</span>
        <el-button size="small" type="warning" plain @click="markersStore.addMode = false">取消</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { createApp } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, Plus } from '@element-plus/icons-vue'
import AMapLoader from '@amap/amap-jsapi-loader'
import * as markersApi from '@/api/markers'
import MarkerInfo from './MarkerInfo.vue'
import MarkerAvatar from './MarkerAvatar.vue'
import ElementPlus from 'element-plus'
import { useMarkersStore } from '@/stores/markers'

const markersStore = useMarkersStore()

const amapKey = import.meta.env.VITE_AMAP_KEY
const map = ref(null)
const loading = ref(false)
const error = ref(null)
let infoWindow = null
let dblClickTimer = null

const initMap = async () => {
  try {
    loading.value = true
    error.value = null

    const AMap = await AMapLoader.load({
      key: amapKey,
      version: '2.0',
      plugins: ['AMap.Scale', 'AMap.Marker', 'AMap.InfoWindow', 'AMap.Geocoder']
    })

    window.AMap = AMap

    map.value = new AMap.Map('map-container', {
      zoom: 5,
      center: [104.0, 35.0],
      mapStyle: 'amap://styles/fresh'
    })

    map.value.addControl(new AMap.Scale())

    map.value.on('click', (e) => {
      if (markersStore.addMode) return
      handleMapClick(e.lnglat)
    })

    const mapContainer = document.getElementById('map-container')
    if (mapContainer) {
      mapContainer.addEventListener('contextmenu', (e) => {
        if (!markersStore.addMode) return
        e.preventDefault()
        e.stopPropagation()
        const pixel = new window.AMap.Pixel(e.clientX, e.clientY)
        const lnglat = map.value.containerToLngLat(pixel)
        handleAddModeRightClick(lnglat)
      })
    }

    await markersStore.fetchMarkers()
    updateMapMarkers(markersStore.markers)

  } catch (err) {
    console.error('地图初始化失败:', err)
    error.value = `地图加载失败: ${err.message}`
  } finally {
    loading.value = false
  }
}

const handleAddModeRightClick = (lnglat) => {
  const lng = lnglat.getLng()
  const lat = lnglat.getLat()

  markersStore.pendingCoords = {
    lng: lng.toFixed(6),
    lat: lat.toFixed(6)
  }
}

const handleMapClick = (lnglat) => {
  const lng = lnglat.getLng()
  const lat = lnglat.getLat()

  const geocoder = new window.AMap.Geocoder()
  geocoder.getAddress([lng, lat], (status, result) => {
    if (status === 'complete' && result.regeocode) {
      const address = result.regeocode.formattedAddress || ''

      const clickContainer = document.createElement('div')
      const clickApp = createApp({
        template: `
          <div class="map-click-card">
            <p class="click-coords">经度: {{ lng.toFixed(6) }}，纬度: {{ lat.toFixed(6) }}</p>
            <p class="click-address" v-if="address">{{ address }}</p>
            <el-button type="primary" size="small" @click="addHere">在此添加足迹</el-button>
          </div>
        `,
        data() {
          return { lng, lat, address }
        },
        methods: {
          addHere() {
            markersStore.setEditingMarker(null)
            markersStore.pendingCoords = {
              lng: lng.toFixed(6),
              lat: lat.toFixed(6)
            }
            if (infoWindow) {
              infoWindow.close()
            }
          }
        }
      })
      clickApp.use(ElementPlus)
      clickApp.mount(clickContainer)

      infoWindow = new window.AMap.InfoWindow({
        content: clickContainer,
        offset: new window.AMap.Pixel(0, -30),
        closeWhenClickMap: true
      })

      infoWindow.open(map.value, [lng, lat])
      infoWindow.on('close', () => clickApp.unmount())
    }
  })
}

const updateMapMarkers = (markersData) => {
  if (!map.value) return
  map.value.clearMap()

  markersData.forEach((marker) => {
    const lng = Number(marker.longitude)
    const lat = Number(marker.latitude)
    if (isNaN(lng) || isNaN(lat)) return

    const container = document.createElement('div')
    const markerApp = createApp(MarkerAvatar, {
      photo: marker.photos && marker.photos.length > 0 ? marker.photos[0] : '',
      category: marker.category || ''
    })
    markerApp.mount(container)

    const amapMarker = new window.AMap.Marker({
      position: [lng, lat],
      map: map.value,
      content: container,
      offset: new window.AMap.Pixel(-15, -15),
      extData: marker
    })

    amapMarker.on('click', (e) => {
      if (infoWindow) {
        infoWindow.close()
      }
      showMarkerInfo(marker)
    })
  })
}

const showMarkerInfo = (marker) => {
  const container = document.createElement('div')
  const infoApp = createApp(MarkerInfo, {
    marker: marker,
    onEdit: () => {
      markersStore.setEditingMarker(marker)
      if (infoWindow) infoWindow.close()
    },
    onDelete: async () => {
      try {
        await markersStore.removeMarker(marker.id)
        ElMessage.success('足迹已删除')
        updateMapMarkers(markersStore.markers)
        if (infoWindow) infoWindow.close()
      } catch (err) {
        ElMessage.error('删除失败: ' + err.message)
      }
    }
  })
  infoApp.mount(container)

  infoWindow = new window.AMap.InfoWindow({
    content: container,
    offset: new window.AMap.Pixel(0, -35),
    closeWhenClickMap: true
  })

  infoWindow.open(map.value, [Number(marker.longitude), Number(marker.latitude)])
  infoWindow.on('close', () => infoApp.unmount())
}

const refreshMap = () => {
  updateMapMarkers(markersStore.markers)
}

const handleSearch = async (params) => {
  try {
    loading.value = true
    const res = await markersApi.searchMarkers(params)
    if (res.data) {
      markersStore.markers = res.data
      updateMapMarkers(res.data)
    }
  } catch (err) {
    ElMessage.error('搜索失败: ' + err.message)
  } finally {
    loading.value = false
  }
}

defineExpose({ refreshMap, handleSearch })

onMounted(() => {
  initMap()
})

onBeforeUnmount(() => {
  if (map.value) {
    map.value.destroy()
    map.value = null
  }
})
</script>

<style scoped>
#map-container {
  width: 100vw;
  height: 100vh;
  position: relative;
}

#map-container.add-mode-cursor {
  cursor: crosshair;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  color: #666;
}

.error-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.error-message {
  text-align: center;
  color: #f56c6c;
}

.error-message p {
  margin-bottom: 12px;
  font-size: 14px;
}

.add-mode-hint-overlay {
  position: absolute;
  bottom: 40px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
  pointer-events: none;
}

.add-mode-hint-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 24px;
  background: rgba(0, 0, 0, 0.75);
  color: #fff;
  border-radius: 24px;
  font-size: 15px;
  backdrop-filter: blur(10px);
  pointer-events: auto;
  animation: fadeInUp 0.3s ease;
}

.add-mode-hint-box .el-icon {
  font-size: 18px;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>

<style>
.map-click-card {
  padding: 10px 14px;
  font-size: 13px;
  min-width: 200px;
}

.map-click-card .click-coords {
  margin: 0 0 4px 0;
  color: #666;
  font-size: 12px;
}

.map-click-card .click-address {
  margin: 0 0 8px 0;
  color: #333;
  font-weight: 500;
}

.map-click-card .el-button {
  width: 100%;
}
</style>