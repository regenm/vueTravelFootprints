<template>
  <div
    class="carousel"
    @mouseenter="paused = true"
    @mouseleave="paused = false"
    @touchstart.passive="onTouchStart"
    @touchend.passive="onTouchEnd"
  >
    <div v-if="!images.length" class="slide placeholder" :style="{ background: placeholderColor }">
      <span>{{ placeholderEmoji }}</span>
    </div>

    <template v-else>
      <div class="viewport" @click="emit('preview', index)">
        <div class="track" :style="{ transform: `translateX(-${index * 100}%)` }">
          <div v-for="(src, i) in images" :key="src + i" class="slide">
            <img :src="src" alt="" @error="onImgError" />
          </div>
        </div>
        <div class="veil"></div>
      </div>

      <button v-if="images.length > 1" class="nav prev" type="button" @click.stop="prev">‹</button>
      <button v-if="images.length > 1" class="nav next" type="button" @click.stop="next">›</button>

      <div v-if="images.length > 1" class="counter">{{ index + 1 }} / {{ images.length }}</div>

      <div v-if="images.length > 1" class="dots">
        <button
          v-for="(_, i) in images"
          :key="i"
          type="button"
          :class="{ active: i === index }"
          @click.stop="index = i"
        />
      </div>
    </template>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { PLACEHOLDER_IMAGE, resolveImageUrl } from '@/utils/image'

const props = defineProps({
  photos: { type: Array, default: () => [] },
  autoplay: { type: Boolean, default: true },
  interval: { type: Number, default: 4200 },
  placeholderEmoji: { type: String, default: '📍' },
  placeholderColor: { type: String, default: '#0f6e6b' }
})

const emit = defineEmits(['preview'])

const index = ref(0)
const paused = ref(false)
const images = ref([])
let timer = null
let startX = 0

watch(
  () => props.photos,
  (list) => {
    images.value = (list || []).filter(Boolean).map(resolveImageUrl)
    if (index.value >= images.value.length) index.value = 0
  },
  { immediate: true }
)

function next() {
  if (images.value.length < 2) return
  index.value = (index.value + 1) % images.value.length
}

function prev() {
  if (images.value.length < 2) return
  index.value = (index.value - 1 + images.value.length) % images.value.length
}

function tick() {
  if (!props.autoplay || paused.value || images.value.length < 2) return
  next()
}

function onTouchStart(e) {
  startX = e.changedTouches[0].clientX
  paused.value = true
}

function onTouchEnd(e) {
  const dx = e.changedTouches[0].clientX - startX
  if (dx > 40) prev()
  else if (dx < -40) next()
  paused.value = false
}

function onImgError(e) {
  e.target.src = PLACEHOLDER_IMAGE
  e.target.onerror = null
}

onMounted(() => {
  timer = setInterval(tick, props.interval)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})

defineExpose({ index, next, prev })
</script>

<style scoped>
.carousel {
  position: relative;
  width: 100%;
  height: 240px;
  overflow: hidden;
  background: #1c1915;
  user-select: none;
}

.viewport {
  width: 100%;
  height: 100%;
  cursor: zoom-in;
}

.track {
  display: flex;
  height: 100%;
  transition: transform 0.45s cubic-bezier(0.22, 1, 0.36, 1);
}

.slide {
  min-width: 100%;
  height: 100%;
}

.slide img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.placeholder {
  display: grid;
  place-items: center;
  height: 100%;
  color: #fff;
  font-size: 48px;
}

.veil {
  position: absolute;
  inset: auto 0 0;
  height: 88px;
  background: linear-gradient(to top, rgba(28, 25, 21, 0.55), transparent);
  pointer-events: none;
}

.nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 50%;
  background: rgba(255, 253, 248, 0.88);
  color: #1c1915;
  font-size: 22px;
  cursor: pointer;
  z-index: 2;
}

.nav.prev { left: 10px; }
.nav.next { right: 10px; }

.counter {
  position: absolute;
  top: 12px;
  right: 12px;
  padding: 3px 10px;
  border-radius: 999px;
  background: rgba(28, 25, 21, 0.55);
  color: #fff;
  font-size: 12px;
  backdrop-filter: blur(6px);
}

.dots {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 12px;
  display: flex;
  justify-content: center;
  gap: 6px;
  z-index: 2;
}

.dots button {
  width: 7px;
  height: 7px;
  border: none;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.45);
  cursor: pointer;
}

.dots button.active {
  width: 18px;
  background: #fff;
}
</style>
