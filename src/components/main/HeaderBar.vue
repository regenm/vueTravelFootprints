<template>
  <div class="header-bar">
    <div class="header-left">
      <el-button
        @click="toggleAddMode"
        :type="markersStore.addMode ? 'warning' : 'primary'"
        :plain="markersStore.addMode"
      >
        <el-icon><Plus /></el-icon>
        {{ markersStore.addMode ? '退出添加模式' : '增加足迹' }}
      </el-button>
      <span v-if="markersStore.addMode" class="add-mode-hint">
        右键地图选择位置
      </span>
    </div>

    <div class="header-right">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索足迹..."
        :prefix-icon="Search"
        clearable
        class="search-input"
        @clear="handleSearchClear"
        @keyup.enter="handleSearch"
      />
      <el-select
        v-model="filterCategory"
        placeholder="分类筛选"
        clearable
        class="filter-select"
        @change="handleFilterChange"
      >
        <el-option
          v-for="cat in markersStore.categoryOptions"
          :key="cat.value"
          :label="cat.label"
          :value="cat.value"
        />
      </el-select>
    </div>

    <el-dialog
      v-model="showDialog"
      :title="isEditing ? '编辑足迹' : '添加新足迹'"
      width="650px"
      :before-close="handleClose"
      destroy-on-close
      append-to-body
    >
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item label="地点名称" prop="name">
          <el-input
            v-model="form.name"
            placeholder="输入地点名称"
          />
        </el-form-item>

        <el-form-item label="到访日期" prop="visitDate">
          <el-date-picker
            v-model="form.visitDate"
            type="date"
            placeholder="选择到访日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="地点分类" prop="category">
          <el-select v-model="form.category" placeholder="选择分类" clearable style="width: 100%">
            <el-option
              v-for="cat in markersStore.categoryOptions"
              :key="cat.value"
              :label="cat.label"
              :value="cat.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="经度" prop="longitude">
          <el-input-number
            v-model="form.longitude"
            :min="-180"
            :max="180"
            :step="0.000001"
            :precision="6"
            placeholder="经度"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="纬度" prop="latitude">
          <el-input-number
            v-model="form.latitude"
            :min="-90"
            :max="90"
            :step="0.000001"
            :precision="6"
            placeholder="纬度"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="旅行笔记" prop="notes">
          <el-input
            v-model="form.notes"
            type="textarea"
            :rows="3"
            placeholder="记录你的旅行故事、美食推荐、小贴士..."
          />
        </el-form-item>

        <el-form-item label="照片">
          <div class="photo-section">
            <div class="photo-upload-row">
              <el-upload
                :auto-upload="false"
                :show-file-list="false"
                accept="image/*"
                :on-change="handleFileChange"
                class="upload-btn"
              >
                <el-button type="primary" plain :loading="uploading">
                  <el-icon><Upload /></el-icon>
                  本地上传
                </el-button>
              </el-upload>
              <span class="upload-hint">或输入图片URL</span>
            </div>

            <div class="photo-urls-container">
              <div
                v-for="(url, index) in form.photos"
                :key="index"
                class="url-input-item"
              >
                <div class="photo-preview" v-if="url">
                  <img :src="url" @error="handleImgError(index)" />
                </div>
                <el-input
                  v-model="form.photos[index]"
                  placeholder="输入图片URL"
                  class="url-input"
                >
                  <template #append>
                    <el-button @click="removePhoto(index)" :disabled="form.photos.length <= 1">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </template>
                </el-input>
              </div>

              <el-button @click="addPhoto" type="primary" plain class="add-url-btn">
                <el-icon><Plus /></el-icon>
                添加图片
              </el-button>

              <div class="url-count">已添加 {{ validPhotoCount }} 张图片</div>
            </div>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showDialog = false">取消</el-button>
          <el-button
            type="primary"
            @click="submitForm"
            :loading="submitting"
            :disabled="!isFormValid"
          >
            {{ submitting ? '提交中...' : (isEditing ? '保存修改' : '确认添加') }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Plus, Upload, Search } from '@element-plus/icons-vue'
import { useMarkersStore } from '@/stores/markers'

const markersStore = useMarkersStore()

const showDialog = ref(false)
const submitting = ref(false)
const uploading = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const formRef = ref(null)
const searchKeyword = ref('')
const filterCategory = ref('')

const form = reactive({
  name: '',
  longitude: null,
  latitude: null,
  visitDate: '',
  category: '',
  notes: '',
  photos: ['']
})

const rules = {
  name: [{ required: true, message: '请输入地点名称', trigger: 'blur' }],
  longitude: [{ required: true, message: '请输入经度', trigger: 'blur' }],
  latitude: [{ required: true, message: '请输入纬度', trigger: 'blur' }]
}

const isFormValid = computed(() => {
  return form.name && form.longitude !== null && form.latitude !== null
})

const validPhotoCount = computed(() => {
  return form.photos.filter(url => url.trim() !== '').length
})

const emit = defineEmits(['marker-added', 'search'])

const toggleAddMode = () => {
  markersStore.addMode = !markersStore.addMode
  if (!markersStore.addMode) {
    markersStore.pendingCoords = null
  }
}

const resetForm = () => {
  form.name = ''
  form.longitude = null
  form.latitude = null
  form.visitDate = ''
  form.category = ''
  form.notes = ''
  form.photos = ['']
  isEditing.value = false
  editingId.value = null
}

const openAddDialog = (coords) => {
  resetForm()
  isEditing.value = false
  if (coords) {
    form.longitude = Number(coords.lng)
    form.latitude = Number(coords.lat)
  }
  showDialog.value = true
}

const openEditDialog = (marker) => {
  isEditing.value = true
  editingId.value = marker.id
  form.name = marker.name
  form.longitude = Number(marker.longitude)
  form.latitude = Number(marker.latitude)
  form.visitDate = marker.visitDate || ''
  form.category = marker.category || ''
  form.notes = marker.notes || ''
  form.photos = marker.photos && marker.photos.length > 0 ? [...marker.photos] : ['']
  showDialog.value = true
}

const handleClose = (done) => {
  if (submitting.value) return
  resetForm()
  done()
}

const addPhoto = () => {
  form.photos.push('')
}

const removePhoto = (index) => {
  if (form.photos.length > 1) {
    form.photos.splice(index, 1)
  }
}

const handleImgError = (index) => {
  form.photos[index] = ''
}

const handleFileChange = async (uploadFile) => {
  uploading.value = true
  try {
    const url = await markersStore.uploadPhoto(uploadFile.raw)
    if (form.photos[0] === '') {
      form.photos[0] = url
    } else {
      form.photos.push(url)
    }
    ElMessage.success('图片上传成功')
  } catch (err) {
    ElMessage.error('图片上传失败: ' + err.message)
  } finally {
    uploading.value = false
  }
}

const submitForm = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    submitting.value = true

    const filteredPhotos = form.photos.filter(url => url.trim() !== '')

    const markerData = {
      name: form.name,
      longitude: String(form.longitude),
      latitude: String(form.latitude),
      photos: filteredPhotos,
      category: form.category,
      notes: form.notes,
      visitDate: form.visitDate
    }

    if (isEditing.value) {
      await markersStore.editMarker(editingId.value, markerData)
      ElMessage.success('足迹修改成功')
    } else {
      await markersStore.addMarker(markerData)
      ElMessage.success('足迹添加成功')
    }

    markersStore.addMode = false
    markersStore.pendingCoords = null
    emit('marker-added', markersStore.markers)
    showDialog.value = false
    resetForm()
  } catch (error) {
    if (error.message) {
      ElMessage.error((isEditing.value ? '修改' : '添加') + '足迹失败: ' + error.message)
    }
  } finally {
    submitting.value = false
  }
}

const handleSearch = () => {
  emit('search', { keyword: searchKeyword.value, category: filterCategory.value })
}

const handleSearchClear = () => {
  emit('search', { keyword: '', category: filterCategory.value })
}

const handleFilterChange = () => {
  emit('search', { keyword: searchKeyword.value, category: filterCategory.value })
}

watch(() => markersStore.editingMarker, (marker) => {
  if (marker) {
    openEditDialog(marker)
  }
})

watch(() => markersStore.pendingCoords, (coords) => {
  if (coords) {
    openAddDialog(coords)
    markersStore.pendingCoords = null
  }
})

defineExpose({ openAddDialog, openEditDialog })
</script>

<style scoped>
.header-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 20px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.header-left {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.add-mode-hint {
  color: #e6a23c;
  font-size: 14px;
  font-weight: 500;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-input {
  width: 220px;
}

.filter-select {
  width: 140px;
}

.photo-section {
  width: 100%;
}

.photo-upload-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.upload-hint {
  color: #909399;
  font-size: 13px;
}

.photo-urls-container {
  width: 100%;
}

.url-input-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.photo-preview {
  width: 48px;
  height: 48px;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid #eee;
  flex-shrink: 0;
}

.photo-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.url-input {
  flex: 1;
}

.add-url-btn {
  margin-top: 4px;
}

.url-count {
  margin-top: 6px;
  color: #909399;
  font-size: 12px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 768px) {
  .header-bar {
    flex-direction: column;
    gap: 10px;
    padding: 8px 12px;
  }

  .header-right {
    width: 100%;
    flex-wrap: wrap;
  }

  .search-input {
    flex: 1;
    min-width: 120px;
  }

  .filter-select {
    width: 120px;
  }
}
</style>