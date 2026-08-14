import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as markersApi from '@/api/markers'

export const useMarkersStore = defineStore('markers', () => {
  const markers = ref([])
  const loading = ref(false)
  const error = ref(null)
  const selectedMarker = ref(null)
  const editingMarker = ref(null)
  const addMode = ref(false)
  const pendingCoords = ref(null)

  const categoryOptions = [
    { label: '🏔️ 自然风光', value: '自然风光' },
    { label: '🏛️ 历史古迹', value: '历史古迹' },
    { label: '🍜 美食探店', value: '美食探店' },
    { label: '🏙️ 城市漫步', value: '城市漫步' },
    { label: '🏖️ 海滩度假', value: '海滩度假' },
    { label: '⛩️ 文化体验', value: '文化体验' },
    { label: '🚗 自驾路书', value: '自驾路书' }
  ]

  const categoryMap = computed(() => {
    const map = {}
    categoryOptions.forEach(c => { map[c.value] = c.label })
    return map
  })

  const filteredMarkers = computed(() => markers.value)

  async function fetchMarkers() {
    loading.value = true
    error.value = null
    try {
      const res = await markersApi.getMarkers()
      markers.value = res.data || []
      return markers.value
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function addMarker(data) {
    const res = await markersApi.createMarker(data)
    markers.value = res.data || []
    return res.data
  }

  async function editMarker(id, data) {
    const res = await markersApi.updateMarker(id, data)
    markers.value = res.data || []
    editingMarker.value = null
    return res.data
  }

  async function removeMarker(id) {
    const res = await markersApi.deleteMarker(id)
    markers.value = res.data || []
    selectedMarker.value = null
    return res.data
  }

  async function uploadPhoto(file) {
    const res = await markersApi.uploadImage(file)
    return res.url
  }

  function setSelectedMarker(marker) {
    selectedMarker.value = marker
  }

  function setEditingMarker(marker) {
    editingMarker.value = marker ? { ...marker } : null
  }

  return {
    markers,
    loading,
    error,
    selectedMarker,
    editingMarker,
    addMode,
    pendingCoords,
    categoryOptions,
    categoryMap,
    filteredMarkers,
    fetchMarkers,
    addMarker,
    editMarker,
    removeMarker,
    uploadPhoto,
    setSelectedMarker,
    setEditingMarker
  }
})