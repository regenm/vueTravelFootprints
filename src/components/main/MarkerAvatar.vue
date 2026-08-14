<template>
  <div class="marker-avatar" :style="{ borderColor: borderColor }">
    <img v-if="photo" :src="photo" alt="marker" />
    <div v-else class="marker-placeholder" :style="{ background: bgColor }">
      {{ categoryEmoji }}
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  photo: { type: String, default: '' },
  category: { type: String, default: '' }
})

const categoryConfig = {
  '自然风光': { color: '#52c41a', emoji: '🏔️' },
  '历史古迹': { color: '#fa8c16', emoji: '🏛️' },
  '美食探店': { color: '#f5222d', emoji: '🍜' },
  '城市漫步': { color: '#1890ff', emoji: '🏙️' },
  '海滩度假': { color: '#13c2c2', emoji: '🏖️' },
  '文化体验': { color: '#722ed1', emoji: '⛩️' },
  '自驾路书': { color: '#eb2f96', emoji: '🚗' }
}

const borderColor = computed(() => {
  return categoryConfig[props.category]?.color || '#409eff'
})

const bgColor = computed(() => {
  return categoryConfig[props.category]?.color || '#409eff'
})

const categoryEmoji = computed(() => {
  return categoryConfig[props.category]?.emoji || '📍'
})
</script>

<style scoped>
.marker-avatar {
  width: 40px;
  height: 40px;
  overflow: hidden;
  border: 3px solid #409eff;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
  background: #fff;
  transition: transform 0.2s;
  cursor: pointer;
}

.marker-avatar:hover {
  transform: scale(1.15);
}

.marker-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.marker-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: #fff;
}
</style>