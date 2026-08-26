<template>
  <el-dialog
    :model-value="ui.shareOpen"
    title="共享旅行地图"
    width="680px"
    append-to-body
    destroy-on-close
    @close="ui.shareOpen = false"
  >
    <el-tabs v-model="tab">
      <el-tab-pane label="创建共享地图" name="create">
        <el-form label-position="top">
          <el-form-item label="标题">
            <el-input v-model="form.title" placeholder="例如：2024 关西赏枫" />
          </el-form-item>
          <el-form-item label="介绍">
            <el-input v-model="form.description" type="textarea" :rows="2" placeholder="给同行的人的一句话" />
          </el-form-item>
          <el-form-item label="分享范围">
            <el-radio-group v-model="form.type">
              <el-radio value="all">我的全部足迹</el-radio>
              <el-radio value="selected">选择部分足迹</el-radio>
            </el-radio-group>
          </el-form-item>
          <div v-if="form.type === 'selected'" class="picker">
            <p class="picker-tip">勾选要放进这张共享地图的地点（{{ selectedIds.length }} 处）</p>
            <label v-for="m in ownMarkers" :key="m.id" class="pick">
              <el-checkbox :model-value="selectedIds.includes(m.id)" @change="toggleMarker(m.id)" />
              <UserAvatar :src="auth.user?.avatar" :name="auth.displayName" :size="22" />
              <span>{{ m.name }}</span>
              <small>{{ m.visitDate || '未注明日期' }}</small>
            </label>
            <p v-if="!ownMarkers.length" class="empty">先去记录一些足迹，再来分享。</p>
          </div>
          <el-form-item label="谁可以打开">
            <el-radio-group v-model="form.isPublic">
              <el-radio :value="true">获得链接即可打开</el-radio>
              <el-radio :value="false">仅受邀成员可打开</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="协作权限">
            <el-radio-group v-model="form.permission">
              <el-radio value="view">仅查看</el-radio>
              <el-radio value="edit">受邀成员可一起记录</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-button type="primary" :loading="creating" @click="create">生成共享地图</el-button>
        </el-form>

        <div v-if="created" class="created">
          <p>地图已创建。复制链接发给朋友，或按用户名邀请已有账号。</p>
          <el-input :model-value="createdLink" readonly>
            <template #append>
              <el-button @click="copy(createdLink)">复制</el-button>
            </template>
          </el-input>
        </div>
      </el-tab-pane>

      <el-tab-pane label="我发出的" name="mine">
        <div v-if="!markers.myShares.length" class="empty">还没有创建过共享地图</div>
        <div v-for="item in markers.myShares" :key="item.id" class="share-row">
          <div class="share-main">
            <div class="share-title">
              <strong>{{ item.title }}</strong>
              <el-tag size="small" :type="item.isPublic ? 'success' : 'info'" effect="plain">
                {{ item.isPublic ? '公开链接' : '仅成员' }}
              </el-tag>
            </div>
            <p>{{ item.markerCount }} 个地点 · {{ item.permission === 'edit' ? '可协作' : '只读' }}</p>
            <div class="faces">
              <UserAvatar
                v-for="p in (item.participants || []).slice(0, 6)"
                :key="p.id"
                :src="p.avatar"
                :name="p.displayName"
                :size="24"
              />
            </div>
            <el-input :model-value="linkOf(item)" size="small" readonly />
            <div class="invite">
              <el-input v-model="invites[item.id]" size="small" placeholder="输入用户名邀请" />
              <el-button size="small" @click="invite(item)">邀请</el-button>
            </div>
            <div v-if="item.members?.length" class="members">
              成员：
              <span v-for="m in item.members" :key="m.userId" class="member">
                <UserAvatar :src="m.avatar" :name="m.displayName" :size="18" />
                {{ m.displayName }}
                <button type="button" title="移除" @click="kick(item, m)">×</button>
              </span>
            </div>
          </div>
          <div class="row-actions">
            <el-button size="small" @click="openShare(item)">打开地图</el-button>
            <el-button size="small" @click="copy(linkOf(item))">复制链接</el-button>
            <el-button size="small" plain @click="togglePublic(item)">
              {{ item.isPublic ? '改为仅成员' : '改为公开' }}
            </el-button>
            <el-button size="small" type="danger" plain @click="revoke(item)">撤销</el-button>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="与我共享" name="inbox">
        <div v-if="!markers.inboxShares.length" class="empty">还没有收到别人的旅行地图</div>
        <div v-for="item in markers.inboxShares" :key="item.id" class="inbox-row">
          <button class="inbox" type="button" @click="openShare(item)">
            <UserAvatar :src="item.owner?.avatar" :name="item.owner?.displayName" :size="40" />
            <span>
              <strong>{{ item.title }}</strong>
              <small>{{ item.owner?.displayName }} · {{ item.markerCount }} 个地点 · {{ item.myPermission === 'edit' ? '可一起记录' : '只读' }}</small>
            </span>
          </button>
          <el-button size="small" plain @click="leave(item)">退出</el-button>
        </div>
      </el-tab-pane>
    </el-tabs>
  </el-dialog>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as markersApi from '@/api/markers'
import * as sharesApi from '@/api/shares'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { useAuthStore } from '@/stores/auth'
import { useMarkersStore } from '@/stores/markers'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
const markers = useMarkersStore()
const auth = useAuthStore()
const router = useRouter()
const tab = ref('create')
const creating = ref(false)
const created = ref(null)
const invites = reactive({})
const ownMarkers = ref([])
const selectedIds = ref([])

const form = reactive({
  title: '我的旅行足迹',
  description: '',
  type: 'all',
  permission: 'view',
  isPublic: true
})

const createdLink = computed(() => (created.value ? linkOf(created.value) : ''))

watch(
  () => ui.shareOpen,
  async (open) => {
    if (!open) return
    form.type = ui.shareTarget ? 'selected' : 'all'
    form.title = ui.shareTarget ? `足迹：${ui.shareTarget.name}` : '我的旅行足迹'
    form.description = ''
    form.permission = 'view'
    form.isPublic = true
    created.value = null
    selectedIds.value = ui.shareTarget?.id ? [ui.shareTarget.id] : []
    try {
      const [res] = await Promise.all([markersApi.getMarkers(), markers.refreshShares()])
      ownMarkers.value = res.data || []
    } catch (err) {
      ElMessage.error(err.message)
    }
  }
)

function linkOf(item) {
  return `${window.location.origin}/s/${item.token}`
}

function toggleMarker(id) {
  const i = selectedIds.value.indexOf(id)
  if (i >= 0) selectedIds.value.splice(i, 1)
  else selectedIds.value.push(id)
}

async function copy(text) {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('链接已复制')
  } catch {
    ElMessage.info(text)
  }
}

async function create() {
  if (form.type === 'selected' && !selectedIds.value.length) {
    ElMessage.warning('请至少选择一条足迹')
    return
  }
  creating.value = true
  try {
    const payload = {
      title: form.title,
      description: form.description,
      type: form.type,
      permission: form.permission,
      isPublic: form.isPublic,
      markerIds: form.type === 'selected' ? selectedIds.value : []
    }
    const res = await sharesApi.createShare(payload)
    created.value = res.data
    await markers.refreshShares()
    ElMessage.success('共享地图已创建')
    tab.value = 'mine'
  } catch (err) {
    ElMessage.error(err.message)
  } finally {
    creating.value = false
  }
}

async function invite(item) {
  const username = (invites[item.id] || '').trim()
  if (!username) return
  try {
    await sharesApi.addShareMember(item.id, { username, permission: item.permission || 'view' })
    invites[item.id] = ''
    await markers.refreshShares()
    ElMessage.success(`已邀请 ${username}`)
  } catch (err) {
    ElMessage.error(err.message)
  }
}

async function kick(item, member) {
  await sharesApi.removeShareMember(item.id, member.userId)
  await markers.refreshShares()
}

async function togglePublic(item) {
  try {
    await sharesApi.updateShare(item.id, { isPublic: !item.isPublic })
    await markers.refreshShares()
    ElMessage.success(item.isPublic ? '已改为仅受邀成员可打开' : '已改为公开链接')
  } catch (err) {
    ElMessage.error(err.message)
  }
}

async function revoke(item) {
  await ElMessageBox.confirm('撤销后，已发出的链接将立即失效。', '撤销分享', { type: 'warning' })
  await sharesApi.deleteShare(item.id)
  await markers.refreshShares()
  ElMessage.success('已撤销')
}

async function leave(item) {
  await ElMessageBox.confirm(`退出「${item.title}」后将无法再打开这张私密地图。`, '退出共享', { type: 'warning' })
  await sharesApi.leaveShare(item.id)
  await markers.refreshShares()
  ElMessage.success('已退出')
}

function openShare(item) {
  ui.shareOpen = false
  router.push(`/s/${item.token}`)
}
</script>

<style scoped>
.created {
  margin-top: 18px;
  padding: 12px;
  background: var(--tf-paper);
  border-radius: 12px;
}

.picker {
  max-height: 220px;
  overflow: auto;
  margin: -4px 0 16px;
  padding: 8px;
  background: var(--tf-paper);
  border-radius: 12px;
}

.picker-tip {
  font-size: 12px;
  color: var(--tf-ink-faint);
  margin-bottom: 8px;
}

.pick {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 4px;
  cursor: pointer;
  border-radius: 8px;
}

.pick:hover {
  background: #fff;
}

.pick span {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pick small {
  color: var(--tf-ink-faint);
  font-size: 12px;
}

.share-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--tf-line);
}

.share-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.share-row p {
  color: var(--tf-ink-faint);
  font-size: 12px;
  margin: 4px 0 8px;
}

.faces {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
}

.share-main {
  flex: 1;
  min-width: 0;
}

.invite {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.members {
  margin-top: 8px;
  font-size: 12px;
  color: var(--tf-ink-soft);
}

.member {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-right: 10px;
}

.members button {
  border: none;
  background: transparent;
  cursor: pointer;
}

.row-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex-shrink: 0;
}

.inbox-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.inbox {
  flex: 1;
  text-align: left;
  border: none;
  background: var(--tf-paper);
  border-radius: 12px;
  padding: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
}

.inbox span {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.inbox small {
  color: var(--tf-ink-faint);
  font-size: 12px;
}

.empty {
  color: var(--tf-ink-faint);
  padding: 20px 0;
}

@media (max-width: 720px) {
  .share-row {
    flex-direction: column;
  }
  .row-actions {
    flex-direction: row;
    flex-wrap: wrap;
  }
}
</style>
