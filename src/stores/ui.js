import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const sidebarOpen = ref(window.innerWidth > 860)
  const detailOpen = ref(false)
  const formOpen = ref(false)
  const shareOpen = ref(false)
  const addMode = ref(false)
  const pendingCoords = ref(null)
  const keyword = ref('')
  const category = ref('')
  const editingMarker = ref(null)
  const shareTarget = ref(null)
  const usersOpen = ref(false)
  const profileOpen = ref(false)
  const focusCoords = ref(null)

  function openForm(marker = null, coords = null) {
    editingMarker.value = marker ? { ...marker } : null
    pendingCoords.value = coords
    addMode.value = false
    formOpen.value = true
  }

  function pickOnMap() {
    formOpen.value = false
    editingMarker.value = null
    pendingCoords.value = null
    addMode.value = true
  }

  function focusMap(lng, lat) {
    focusCoords.value = { lng: Number(lng), lat: Number(lat), at: Date.now() }
  }

  function closeForm() {
    formOpen.value = false
    editingMarker.value = null
    pendingCoords.value = null
  }

  function toggleAddMode() {
    addMode.value = !addMode.value
    if (!addMode.value) pendingCoords.value = null
  }

  function openShare(target = null) {
    shareTarget.value = target
    shareOpen.value = true
  }

  return {
    sidebarOpen,
    detailOpen,
    formOpen,
    shareOpen,
    addMode,
    pendingCoords,
    keyword,
    category,
    editingMarker,
    shareTarget,
    usersOpen,
    profileOpen,
    focusCoords,
    openForm,
    closeForm,
    pickOnMap,
    focusMap,
    toggleAddMode,
    openShare
  }
})
