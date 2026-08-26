<template>
  <div v-if="!ready" class="boot-screen">
    <div class="boot-mark">迹</div>
    <p>正在打开你的旅行地图…</p>
  </div>
  <router-view v-else />
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const ready = ref(false)

onMounted(async () => {
  if (auth.token) {
    try {
      await auth.hydrate()
    } catch {
      auth.logout()
    }
  }
  ready.value = true
})
</script>
