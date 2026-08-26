<template>
  <aside v-if="marker" class="detail">
    <button class="close" type="button" @click="close">×</button>

    <ImageCarousel
      :photos="marker.photos"
      :placeholder-emoji="meta.emoji"
      :placeholder-color="meta.color"
      @preview="openPreview"
    />

    <div class="body">
      <div class="kicker">
        <span class="tag">{{ meta.emoji }} {{ meta.label }}</span>
        <span class="date">{{ formatVisitDate(marker.visitDate) }}</span>
      </div>
      <h2>{{ marker.name }}</h2>
      <p v-if="marker.address" class="addr">{{ marker.address }}</p>
      <p v-if="marker.author" class="author">
        <UserAvatar :src="marker.author.avatar" :name="marker.author.displayName" :size="22" />
        {{ marker.author.displayName }}
      </p>

      <div v-if="marker.notes" class="notes">{{ marker.notes }}</div>

      <div v-if="marker.photos?.length > 1" class="thumbs">
        <button
          v-for="(p, i) in marker.photos"
          :key="p + i"
          type="button"
          @click="openPreview(i)"
        >
          <img :src="resolveImageUrl(p)" alt="" />
        </button>
      </div>

      <div class="actions">
        <el-button v-if="canShare" plain @click="ui.openShare(marker)">分享这条</el-button>
        <el-button v-if="canMutate" type="primary" plain @click="ui.openForm(marker)">编辑</el-button>
        <el-button v-if="canMutate" type="danger" plain @click="onDelete">删除</el-button>
      </div>
    </div>

    <el-image-viewer
      v-if="previewOn"
      :url-list="previewList"
      :initial-index="previewIndex"
      teleported
      @close="previewOn = false"
    />
  </aside>
</template>

<script setup>
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useMarkersStore } from '@/stores/markers'
import { useUiStore } from '@/stores/ui'
import { getCategoryMeta } from '@/utils/constants'
import { formatVisitDate } from '@/utils/format'
import { resolveImageUrl } from '@/utils/image'
import ImageCarousel from '@/components/common/ImageCarousel.vue'
import UserAvatar from '@/components/common/UserAvatar.vue'

const markers = useMarkersStore()
const ui = useUiStore()
const auth = useAuthStore()

const marker = computed(() => markers.selectedMarker)
const meta = computed(() => getCategoryMeta(marker.value?.category))
const canMutate = computed(() => marker.value && marker.value.userId === auth.user?.id)
const canShare = computed(() => auth.isLoggedIn && canMutate.value)

const previewOn = ref(false)
const previewIndex = ref(0)
const previewList = computed(() => (marker.value?.photos || []).map(resolveImageUrl))

function close() {
  ui.detailOpen = false
}

function openPreview(i) {
  previewIndex.value = i || 0
  if (previewList.value.length) previewOn.value = true
}

async function onDelete() {
  try {
    await ElMessageBox.confirm(`确定删除「${marker.value.name}」？删除后无法恢复。`, '删除足迹', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
    await markers.removeMarker(marker.value.id)
    ElMessage.success('已删除')
  } catch (err) {
    if (err !== 'cancel' && err?.message) ElMessage.error(err.message)
  }
}
</script>

<style scoped>
.detail {
  width: var(--detail-w);
  max-width: 100%;
  background: var(--tf-card);
  border-left: 1px solid var(--tf-line);
  display: flex;
  flex-direction: column;
  min-height: 0;
  position: relative;
  z-index: 13;
  animation: slide 0.28s ease;
}

@keyframes slide {
  from { transform: translateX(24px); opacity: 0; }
  to { transform: none; opacity: 1; }
}

.close {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 3;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 50%;
  background: rgba(255, 253, 248, 0.9);
  cursor: pointer;
  font-size: 18px;
}

.body {
  padding: 18px 18px 24px;
  overflow: auto;
}

.kicker {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

.tag {
  background: var(--tf-teal-soft);
  color: var(--tf-teal-deep);
  border-radius: 999px;
  padding: 2px 10px;
  font-size: 12px;
}

.date {
  color: var(--tf-ink-faint);
  font-size: 12px;
}

h2 {
  font-family: var(--font-serif);
  font-size: 28px;
  line-height: 1.2;
  margin-bottom: 8px;
}

.addr,
.author {
  color: var(--tf-ink-soft);
  font-size: 13px;
  margin-bottom: 4px;
}

.author {
  display: flex;
  align-items: center;
  gap: 8px;
}

.notes {
  margin: 14px 0;
  padding: 14px;
  background: var(--tf-paper);
  border-radius: 14px;
  white-space: pre-wrap;
  color: var(--tf-ink-soft);
  font-size: 14px;
  line-height: 1.75;
}

.thumbs {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.thumbs button {
  border: none;
  padding: 0;
  height: 56px;
  border-radius: 10px;
  overflow: hidden;
  cursor: pointer;
}

.thumbs img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

@media (max-width: 860px) {
  .detail {
    position: absolute;
    inset: auto 0 0;
    width: 100%;
    max-height: 78%;
    border-left: none;
    border-top-left-radius: 20px;
    border-top-right-radius: 20px;
    box-shadow: var(--tf-shadow);
    overflow: auto;
  }
}
</style>
