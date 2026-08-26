<template>
  <span
    class="ua"
    :class="{ ring }"
    :style="{ width: `${size}px`, height: `${size}px`, fontSize: `${Math.max(11, size * 0.42)}px`, background: tone }"
    :title="name"
  >
    <img v-if="src && !broken" :src="resolved" alt="" @error="broken = true" />
    <span v-else>{{ letter }}</span>
  </span>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { avatarTone, initials } from '@/utils/format'
import { resolveImageUrl } from '@/utils/image'

const props = defineProps({
  src: { type: String, default: '' },
  name: { type: String, default: '' },
  size: { type: Number, default: 32 },
  ring: { type: Boolean, default: false }
})

const broken = ref(false)
const resolved = computed(() => resolveImageUrl(props.src))
const letter = computed(() => initials(props.name))
const tone = computed(() => avatarTone(props.name || '旅'))

watch(
  () => props.src,
  () => {
    broken.value = false
  }
)
</script>

<style scoped>
.ua {
  display: inline-grid;
  place-items: center;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  color: #fff;
  font-weight: 700;
  line-height: 1;
  user-select: none;
  vertical-align: middle;
}

.ua.ring {
  box-shadow: 0 0 0 2px #fffdf8;
}

.ua img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
</style>
