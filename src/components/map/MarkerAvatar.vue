<template>
  <button class="pin" type="button" :class="{ active }" :style="{ '--pin': color }" :title="title">
    <span class="face">
      <img v-if="cover" :src="cover" alt="" />
      <span v-else class="emoji">{{ emoji }}</span>
    </span>
    <span v-if="showAuthor" class="who">
      <UserAvatar :src="authorAvatar" :name="authorName" :size="18" ring />
    </span>
  </button>
</template>

<script setup>
import { computed } from 'vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { getCategoryMeta } from '@/utils/constants'
import { resolveImageUrl } from '@/utils/image'

const props = defineProps({
  photo: { type: String, default: '' },
  category: { type: String, default: '' },
  active: { type: Boolean, default: false },
  showAuthor: { type: Boolean, default: false },
  authorAvatar: { type: String, default: '' },
  authorName: { type: String, default: '' }
})

const meta = computed(() => getCategoryMeta(props.category))
const color = computed(() => meta.value.color)
const emoji = computed(() => meta.value.emoji)
const cover = computed(() => resolveImageUrl(props.photo))
const title = computed(() => (props.showAuthor && props.authorName ? props.authorName : ''))
</script>

<style scoped>
.pin {
  position: relative;
  width: 42px;
  height: 42px;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  transform: rotate(-8deg);
  transition: transform 0.2s ease;
}

.face {
  display: block;
  width: 42px;
  height: 42px;
  border: 3px solid var(--pin, #0f6e6b);
  border-radius: 14px 14px 14px 4px;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 8px 18px rgba(28, 25, 21, 0.28);
}

.pin img,
.emoji {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: grid;
  place-items: center;
  font-size: 18px;
}

.pin:hover,
.pin.active {
  transform: rotate(0deg) scale(1.12);
  z-index: 2;
}

.who {
  position: absolute;
  right: -7px;
  bottom: -7px;
  transform: rotate(8deg);
  pointer-events: none;
}
</style>
