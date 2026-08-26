import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as markersApi from '@/api/markers'
import * as sharesApi from '@/api/shares'
import { CATEGORIES, getCategoryMeta } from '@/utils/constants'
import { compressImage } from '@/utils/image'
import { useUiStore } from './ui'

export const useMarkersStore = defineStore('markers', () => {
  const markers = ref([])
  const loading = ref(false)
  const selectedId = ref(null)
  const source = ref('mine')
  const currentShare = ref(null)
  const myShares = ref([])
  const inboxShares = ref([])

  const selectedMarker = computed(() => markers.value.find((m) => m.id === selectedId.value) || null)
  const canEdit = computed(() => source.value === 'mine' || Boolean(currentShare.value?.canEdit))
  const isShareView = computed(() => source.value === 'share')

  const filteredMarkers = computed(() => {
    const ui = useUiStore()
    const kw = ui.keyword.trim().toLowerCase()
    const cat = ui.category
    return markers.value.filter((m) => {
      const hitCat = !cat || m.category === cat
      if (!hitCat) return false
      if (!kw) return true
      return [m.name, m.notes, m.address, m.category].some((v) => String(v || '').toLowerCase().includes(kw))
    })
  })

  const stats = computed(() => {
    const list = markers.value
    const photos = list.reduce((n, m) => n + (m.photos?.length || 0), 0)
    const cats = new Set(list.map((m) => m.category).filter(Boolean))
    return {
      places: list.length,
      photos,
      categories: cats.size
    }
  })

  function selectMarker(marker, openDetail = true) {
    const ui = useUiStore()
    selectedId.value = marker?.id || null
    ui.detailOpen = Boolean(marker && openDetail)
  }

  async function fetchMarkers() {
    loading.value = true
    try {
      const res = await markersApi.getMarkers()
      source.value = 'mine'
      currentShare.value = null
      markers.value = res.data || []
      return markers.value
    } finally {
      loading.value = false
    }
  }

  async function loadShare(token) {
    loading.value = true
    try {
      const res = await sharesApi.getShareByToken(token)
      source.value = 'share'
      currentShare.value = res.data
      markers.value = res.data?.markers || []
      return res.data
    } finally {
      loading.value = false
    }
  }

  async function addMarker(data) {
    const payload = { ...data }
    if (source.value === 'share' && currentShare.value?.share?.id) {
      payload.shareId = currentShare.value.share.id
    }
    const res = await markersApi.createMarker(payload)
    if (source.value === 'share' && currentShare.value?.share?.token) {
      await loadShare(currentShare.value.share.token)
    } else {
      await fetchMarkers()
    }
    if (res.data?.id) selectMarker(res.data)
    return res.data
  }

  async function editMarker(id, data) {
    const res = await markersApi.updateMarker(id, data)
    const idx = markers.value.findIndex((m) => m.id === id)
    if (idx >= 0) markers.value[idx] = res.data
    else await fetchMarkers()
    return res.data
  }

  async function removeMarker(id) {
    await markersApi.deleteMarker(id)
    markers.value = markers.value.filter((m) => m.id !== id)
    if (selectedId.value === id) {
      selectedId.value = null
      useUiStore().detailOpen = false
    }
  }

  async function uploadPhoto(file) {
    const compressed = await compressImage(file)
    const res = await markersApi.uploadImage(compressed)
    return res.url
  }

  async function refreshShares() {
    const [mine, inbox] = await Promise.all([sharesApi.listMyShares(), sharesApi.listInboxShares()])
    myShares.value = mine.data || []
    inboxShares.value = inbox.data || []
  }

  return {
    markers,
    loading,
    selectedId,
    selectedMarker,
    source,
    currentShare,
    myShares,
    inboxShares,
    canEdit,
    isShareView,
    filteredMarkers,
    stats,
    categoryOptions: CATEGORIES,
    getCategoryMeta,
    selectMarker,
    fetchMarkers,
    loadShare,
    addMarker,
    editMarker,
    removeMarker,
    uploadPhoto,
    refreshShares
  }
})
