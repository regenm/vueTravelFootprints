<template>
  <el-dialog
    :model-value="ui.formOpen"
    :title="ui.editingMarker ? '编辑足迹' : '记录一处足迹'"
    width="720px"
    class="marker-form"
    destroy-on-close
    append-to-body
    @close="ui.closeForm()"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <el-form-item v-if="!ui.editingMarker" label="搜索地点">
        <el-autocomplete
          v-model="placeQuery"
          :fetch-suggestions="queryPlaces"
          placeholder="输入地名，例如：西湖、故宫、浅草寺"
          clearable
          style="width: 100%"
          popper-class="place-suggest"
          :trigger-on-focus="false"
          @select="onPlaceSelect"
        >
          <template #default="{ item }">
            <div class="suggest">
              <strong>{{ item.name }}</strong>
              <small>{{ item.address }}</small>
            </div>
          </template>
        </el-autocomplete>
        <p class="search-tip">
          选中地点后会自动填入名称、地址和坐标。
          <button type="button" class="link" @click="ui.pickOnMap()">改在地图上点选</button>
        </p>
      </el-form-item>

      <div class="grid">
        <el-form-item label="地点名称" prop="name">
          <el-input v-model="form.name" placeholder="例如：西湖、京都清水寺" />
        </el-form-item>
        <el-form-item label="到访日期">
          <el-date-picker
            v-model="form.visitDate"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="选择日期"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.category" clearable placeholder="选择分类" style="width: 100%">
            <el-option
              v-for="cat in markers.categoryOptions"
              :key="cat.value"
              :label="`${cat.emoji} ${cat.label}`"
              :value="cat.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="form.address" placeholder="搜索地点后会自动填充" />
        </el-form-item>
      </div>

      <div class="grid coords">
        <el-form-item label="经度" prop="longitude">
          <el-input-number v-model="form.longitude" :min="-180" :max="180" :precision="6" :step="0.000001" style="width: 100%" />
        </el-form-item>
        <el-form-item label="纬度" prop="latitude">
          <el-input-number v-model="form.latitude" :min="-90" :max="90" :precision="6" :step="0.000001" style="width: 100%" />
        </el-form-item>
      </div>

      <el-form-item label="旅行笔记">
        <el-input
          v-model="form.notes"
          type="textarea"
          :rows="4"
          placeholder="写下那天的天气、食物、路上遇到的人或一句想留给以后的话"
        />
      </el-form-item>

      <el-form-item label="照片">
        <div class="dropzone" @dragover.prevent @drop.prevent="onDrop">
          <el-upload :auto-upload="false" :show-file-list="false" accept="image/*" multiple :on-change="onFile">
            <el-button type="primary" plain :loading="uploading">上传照片</el-button>
          </el-upload>
          <span>可一次选择多张，也可拖到这里。第一张会作为封面。</span>
        </div>
        <div class="photos">
          <div v-for="(url, i) in form.photos" :key="url + i" class="photo">
            <img :src="url" alt="" />
            <em v-if="i === 0">封面</em>
            <button type="button" @click="form.photos.splice(i, 1)">×</button>
          </div>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="ui.closeForm()">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">
        {{ ui.editingMarker ? '保存修改' : '加入地图' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useMarkersStore } from '@/stores/markers'
import { useUiStore } from '@/stores/ui'
import { searchPlaces } from '@/utils/places'

const markers = useMarkersStore()
const ui = useUiStore()
const formRef = ref(null)
const submitting = ref(false)
const uploading = ref(false)
const placeQuery = ref('')
const searching = ref(false)

const form = reactive({
  name: '',
  longitude: null,
  latitude: null,
  address: '',
  visitDate: '',
  category: '',
  notes: '',
  photos: []
})

const rules = {
  name: [{ required: true, message: '请填写地点名称', trigger: 'blur' }],
  longitude: [{ required: true, message: '请搜索地点或填写经度', trigger: 'change' }],
  latitude: [{ required: true, message: '请搜索地点或填写纬度', trigger: 'change' }]
}

function reset() {
  Object.assign(form, {
    name: '',
    longitude: null,
    latitude: null,
    address: '',
    visitDate: '',
    category: '',
    notes: '',
    photos: []
  })
  placeQuery.value = ''
}

watch(
  () => ui.formOpen,
  (open) => {
    if (!open) return
    const editing = ui.editingMarker
    const coords = ui.pendingCoords
    if (editing) {
      Object.assign(form, {
        name: editing.name,
        longitude: Number(editing.longitude),
        latitude: Number(editing.latitude),
        address: editing.address || '',
        visitDate: editing.visitDate || '',
        category: editing.category || '',
        notes: editing.notes || '',
        photos: [...(editing.photos || [])]
      })
      placeQuery.value = editing.name || ''
    } else {
      reset()
      if (coords) {
        form.longitude = Number(coords.lng)
        form.latitude = Number(coords.lat)
        form.address = coords.address || ''
        if (coords.name) form.name = coords.name
        placeQuery.value = coords.name || ''
      }
    }
  }
)

async function queryPlaces(query, cb) {
  const q = String(query || '').trim()
  if (q.length < 2) {
    cb([])
    return
  }
  searching.value = true
  try {
    cb(await searchPlaces(q))
  } catch (err) {
    cb([])
    ElMessage.error(err.message || '地点搜索失败')
  } finally {
    searching.value = false
  }
}

function onPlaceSelect(item) {
  const lng = Number(item.lng)
  const lat = Number(item.lat)
  if (!Number.isFinite(lng) || !Number.isFinite(lat)) {
    ElMessage.error('这个结果没有坐标，请换一个地点')
    return
  }
  form.name = item.name || form.name
  form.address = item.address || form.address
  form.longitude = lng
  form.latitude = lat
  placeQuery.value = item.name || item.value
  ui.focusMap(lng, lat)
  ElMessage.success(`已定位到${form.name}`)
}

async function uploadOne(raw) {
  const url = await markers.uploadPhoto(raw)
  form.photos.push(url)
}

async function onFile(uploadFile) {
  if (!uploadFile?.raw) return
  uploading.value = true
  try {
    await uploadOne(uploadFile.raw)
  } catch (err) {
    ElMessage.error(err.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

async function onDrop(e) {
  const files = [...(e.dataTransfer?.files || [])].filter((f) => f.type.startsWith('image/'))
  if (!files.length) return
  uploading.value = true
  try {
    for (const file of files) await uploadOne(file)
  } catch (err) {
    ElMessage.error(err.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

async function submit() {
  await formRef.value.validate()
  submitting.value = true
  try {
    const payload = {
      name: form.name,
      longitude: String(form.longitude),
      latitude: String(form.latitude),
      address: form.address,
      visitDate: form.visitDate,
      category: form.category,
      notes: form.notes,
      photos: form.photos
    }
    if (ui.editingMarker) {
      await markers.editMarker(ui.editingMarker.id, payload)
      ElMessage.success('足迹已更新')
    } else {
      await markers.addMarker(payload)
      ElMessage.success('足迹已加入地图')
    }
    ui.closeForm()
  } catch (err) {
    if (err?.message) ElMessage.error(err.message)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 16px;
}

.search-tip {
  margin-top: 6px;
  color: var(--tf-ink-faint);
  font-size: 12px;
}

.link {
  border: none;
  background: none;
  color: var(--tf-teal);
  cursor: pointer;
  padding: 0;
}

.suggest {
  display: flex;
  flex-direction: column;
  line-height: 1.3;
  padding: 4px 0;
}

.suggest small {
  color: #8a8176;
  font-size: 12px;
}

.dropzone {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px dashed var(--tf-line);
  border-radius: 12px;
  background: var(--tf-paper);
  color: var(--tf-ink-faint);
  font-size: 13px;
  margin-bottom: 10px;
}

.photos {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.photo {
  position: relative;
  width: 84px;
  height: 64px;
  border-radius: 10px;
  overflow: hidden;
}

.photo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.photo em {
  position: absolute;
  left: 4px;
  bottom: 4px;
  font-style: normal;
  font-size: 10px;
  background: rgba(15, 110, 107, 0.9);
  color: #fff;
  border-radius: 6px;
  padding: 1px 6px;
}

.photo button {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  cursor: pointer;
}

@media (max-width: 720px) {
  .grid { grid-template-columns: 1fr; }
}
</style>

<style>
.place-suggest {
  z-index: 5000 !important;
}
</style>
