<template>
  <div>
    <HeaderBar ref="headerBarRef" @marker-added="onMarkerAdded" @search="onSearch" />
    <MapPart ref="mapPartRef" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import HeaderBar from '@/components/main/HeaderBar.vue'
import MapPart from '@/components/main/MapPart.vue'

const headerBarRef = ref(null)
const mapPartRef = ref(null)

const onMarkerAdded = (markers) => {
  if (mapPartRef.value) {
    mapPartRef.value.refreshMap()
  }
}

const onSearch = (params) => {
  if (mapPartRef.value) {
    mapPartRef.value.handleSearch(params)
  }
}
</script>