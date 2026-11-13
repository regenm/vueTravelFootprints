<template>
  <el-dialog :model-value="ui.tripsOpen" title="行程管理" width="560px" destroy-on-close @close="ui.tripsOpen = false">
    <div class="new-trip">
      <el-input v-model="newName" placeholder="新行程名称，例如：2026 春节云南行" maxlength="30" />
      <el-date-picker
        v-model="newRange"
        type="daterange"
        value-format="YYYY-MM-DD"
        range-separator="至"
        start-placeholder="开始"
        end-placeholder="结束"
      />
      <el-button type="primary" :loading="submitting" @click="create">新建</el-button>
    </div>

    <el-table :data="trips.trips" empty-text="还没有行程">
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column label="日期" min-width="180">
        <template #default="{ row }">{{ row.startDate || row.endDate ? `${row.startDate || '?'} ~ ${row.endDate || '?'}` : '未注明' }}</template>
      </el-table-column>
      <el-table-column label="足迹" width="70">
        <template #default="{ row }">{{ countOf(row.id) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="130">
        <template #default="{ row }">
          <el-button link type="primary" @click="startRename(row)">改名</el-button>
          <el-popconfirm title="删除行程不会删除足迹，仅解除归属。确定？" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <div v-if="renaming" class="rename-row">
      <el-input v-model="renameValue" maxlength="30" @keyup.enter="confirmRename" />
      <el-button type="primary" size="small" @click="confirmRename">保存</el-button>
      <el-button size="small" @click="renaming = null">取消</el-button>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useTripsStore } from '@/stores/trips'
import { useUiStore } from '@/stores/ui'
import { useMarkersStore } from '@/stores/markers'

const ui = useUiStore()
const trips = useTripsStore()
const markers = useMarkersStore()

const newName = ref('')
const newRange = ref(null)
const submitting = ref(false)
const renaming = ref(null)
const renameValue = ref('')

watch(
  () => ui.tripsOpen,
  (open) => {
    if (open) trips.fetchTrips().catch(() => {})
  }
)

function countOf(tripId) {
  return markers.markers.filter((m) => m.tripId === tripId).length
}

async function create() {
  const name = newName.value.trim()
  if (!name) {
    ElMessage.warning('请填写行程名称')
    return
  }
  submitting.value = true
  try {
    await trips.createTrip({
      name,
      startDate: newRange.value?.[0] || '',
      endDate: newRange.value?.[1] || ''
    })
    newName.value = ''
    newRange.value = null
    ElMessage.success('行程已创建')
  } catch (err) {
    ElMessage.error(err.message || '创建失败')
  } finally {
    submitting.value = false
  }
}

function startRename(row) {
  renaming.value = row
  renameValue.value = row.name
}

async function confirmRename() {
  const name = renameValue.value.trim()
  if (!name || !renaming.value) return
  try {
    await trips.updateTrip(renaming.value.id, { name })
    renaming.value = null
    ElMessage.success('已改名')
  } catch (err) {
    ElMessage.error(err.message || '改名失败')
  }
}

async function remove(row) {
  try {
    await trips.removeTrip(row.id)
    if (markers.isShareView) {
      // 分享视图下不能 fetchMarkers（会把地图切回本人足迹），本地同步即可：后端删除行程时已清空足迹的 trip_id
      markers.markers.forEach((m) => {
        if (m.tripId === row.id) m.tripId = ''
      })
    } else {
      await markers.fetchMarkers()
    }
    ElMessage.success('行程已删除，足迹保留')
  } catch (err) {
    ElMessage.error(err.message || '删除失败')
  }
}
</script>

<style scoped>
.new-trip {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
}

.rename-row {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}
</style>
