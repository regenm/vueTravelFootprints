<template>
  <div class="info-card">
    <div v-if="markerData.photos && markerData.photos.length > 0" class="cover-photo">
      <img :src="getImageUrl(markerData.photos[0])" :alt="markerData.name" @error="handleImageError" />
      <div v-if="markerData.photos.length > 1" class="photo-count">
        <el-icon><Picture /></el-icon>
        {{ markerData.photos.length }}
      </div>
    </div>

    <div class="info-body">
      <div class="info-header">
        <h3>{{ markerData.name }}</h3>
        <span v-if="categoryLabel" class="category-tag">{{ categoryLabel }}</span>
      </div>

      <div v-if="markerData.visitDate" class="visit-date">
        <el-icon><Calendar /></el-icon>
        <span>{{ markerData.visitDate }}</span>
      </div>

      <div v-if="markerData.notes" class="notes-section">
        <p>{{ markerData.notes }}</p>
      </div>

      <div v-if="markerData.photos && markerData.photos.length > 1" class="photos-gallery">
        <div
          v-for="(p, i) in markerData.photos"
          :key="i"
          class="gallery-item"
          @click="handleImageClick(i)"
        >
          <img :src="getImageUrl(p)" :alt="`${markerData.name} ${i + 1}`" @error="handleImageError" />
        </div>
      </div>

      <div class="info-actions">
        <el-button class="action-btn edit-btn" @click="handleEdit">
          <el-icon><Edit /></el-icon>
          编辑足迹
        </el-button>
        <el-button class="action-btn delete-btn" @click="handleDelete">
          <el-icon><Delete /></el-icon>
          删除足迹
        </el-button>
      </div>
    </div>

    <el-image-viewer
      v-if="previewVisible"
      :url-list="previewList"
      :initial-index="previewIndex"
      @close="previewVisible = false"
      teleported
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElImageViewer, ElMessageBox } from 'element-plus'
import { Edit, Delete, Calendar, Picture } from '@element-plus/icons-vue'
import { useMarkersStore } from '@/stores/markers'

const props = defineProps({
  marker: { type: Object, default: () => ({}) }
})

const emit = defineEmits(['edit', 'delete', 'close'])

const markersStore = useMarkersStore()

const previewVisible = ref(false)
const previewIndex = ref(0)

const markerData = computed(() => props.marker || {})

const categoryLabel = computed(() => {
  return markersStore.categoryMap[markerData.value.category] || ''
})

const previewList = computed(() => {
  return (markerData.value.photos || []).map(p => getImageUrl(p))
})

function getImageUrl(path) {
  if (!path) return ''
  if (path.startsWith('http') || path.startsWith('data:')) return path
  return path
}

function handleImageError(e) {
  e.target.src = 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 60"><rect fill="%23f0f0f0" width="100" height="60"/><text x="50" y="35" text-anchor="middle" fill="%23bbb" font-size="12">无图片</text></svg>'
  e.target.onerror = null
}

function handleImageClick(index) {
  previewIndex.value = index
  previewVisible.value = true
}

function handleEdit() {
  emit('edit', markerData.value)
}

function handleDelete() {
  ElMessageBox.confirm(`确定要删除"${markerData.value.name}"这个足迹吗？此操作不可恢复。`, '确认删除', {
    confirmButtonText: '确定删除',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    emit('delete', markerData.value)
  }).catch(() => {})
}
</script>

<style scoped>
.info-card {
  width: 340px;
  max-width: 85vw;
  background: #fff;
  border-radius: 14px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.cover-photo {
  position: relative;
  width: 100%;
  height: 160px;
  overflow: hidden;
  background: #f5f5f5;
}

.cover-photo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.photo-count {
  position: absolute;
  bottom: 10px;
  right: 10px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  border-radius: 12px;
  font-size: 12px;
  backdrop-filter: blur(4px);
}

.photo-count .el-icon {
  font-size: 14px;
}

.info-body {
  padding: 16px;
}

.info-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  flex-wrap: wrap;
}

.info-header h3 {
  margin: 0;
  color: #1a1a1a;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.3px;
}

.category-tag {
  display: inline-block;
  padding: 3px 10px;
  background: linear-gradient(135deg, #ecf5ff, #d9ecff);
  color: #409eff;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.visit-date {
  display: flex;
  align-items: center;
  gap: 5px;
  color: #909399;
  font-size: 13px;
  margin-bottom: 10px;
}

.visit-date .el-icon {
  font-size: 15px;
}

.notes-section {
  margin-bottom: 12px;
  padding: 12px;
  background: #fafbfc;
  border-radius: 10px;
  border-left: 3px solid #409eff;
}

.notes-section p {
  margin: 0;
  color: #555;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.photos-gallery {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 14px;
}

.gallery-item {
  width: 72px;
  height: 54px;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.gallery-item:hover {
  border-color: #409eff;
  transform: scale(1.06);
}

.gallery-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.info-actions {
  display: flex;
  gap: 10px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.action-btn {
  flex: 1;
  height: 38px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  transition: all 0.25s;
}

.edit-btn {
  background: #ecf5ff;
  border: 1px solid #d9ecff;
  color: #409eff;
}

.edit-btn:hover {
  background: #409eff;
  border-color: #409eff;
  color: #fff;
}

.delete-btn {
  background: #fef0f0;
  border: 1px solid #fde2e2;
  color: #f56c6c;
}

.delete-btn:hover {
  background: #f56c6c;
  border-color: #f56c6c;
  color: #fff;
}

@media (max-width: 768px) {
  .info-card {
    width: calc(100vw - 32px);
  }

  .cover-photo {
    height: 130px;
  }
}
</style>