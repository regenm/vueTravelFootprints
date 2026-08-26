<template>
  <el-dialog
    :model-value="ui.usersOpen"
    title="用户管理"
    width="680px"
    append-to-body
    destroy-on-close
    @close="ui.usersOpen = false"
  >
    <p class="hint">账号只能由管理员创建，不能自行注册。</p>

    <el-form :model="form" label-position="top" class="create">
      <div class="grid">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" placeholder="3-20 位字母数字或下划线" />
        </el-form-item>
        <el-form-item label="邮箱" required>
          <el-input v-model="form.email" placeholder="user@example.com" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="form.displayName" placeholder="可选" />
        </el-form-item>
        <el-form-item label="初始密码" required>
          <el-input v-model="form.password" type="password" show-password placeholder="至少 8 位" />
        </el-form-item>
      </div>
      <el-form-item label="角色">
        <el-radio-group v-model="form.role">
          <el-radio label="user">普通用户</el-radio>
          <el-radio label="admin">管理员</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-button type="primary" :loading="creating" @click="create">添加用户</el-button>
    </el-form>

    <el-table :data="users" stripe empty-text="还没有其他用户" class="table">
      <el-table-column label="" width="56">
        <template #default="{ row }">
          <UserAvatar :src="row.avatar" :name="row.displayName || row.username" :size="32" />
        </template>
      </el-table-column>
      <el-table-column prop="username" label="用户名" width="140" />
      <el-table-column prop="displayName" label="昵称" />
      <el-table-column prop="role" label="角色" width="100">
        <template #default="{ row }">{{ row.role === 'admin' ? '管理员' : '普通用户' }}</template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180">
        <template #default="{ row }">{{ String(row.createdAt || '').slice(0, 10) }}</template>
      </el-table-column>
    </el-table>
  </el-dialog>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import * as authApi from '@/api/auth'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { useUiStore } from '@/stores/ui'

const ui = useUiStore()
const users = ref([])
const creating = ref(false)
const form = reactive({
  username: '',
  email: '',
  displayName: '',
  password: '',
  role: 'user'
})

watch(
  () => ui.usersOpen,
  async (open) => {
    if (!open) return
    try {
      const res = await authApi.listUsers()
      users.value = res.data || []
    } catch (err) {
      ElMessage.error(err.message || '无法加载用户')
    }
  }
)

async function create() {
  creating.value = true
  try {
    await authApi.createUser({ ...form })
    ElMessage.success(`已创建用户 ${form.username}`)
    form.username = ''
    form.email = ''
    form.displayName = ''
    form.password = ''
    form.role = 'user'
    const res = await authApi.listUsers()
    users.value = res.data || []
  } catch (err) {
    ElMessage.error(err.message || '创建失败')
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.hint {
  color: var(--tf-ink-soft);
  margin-bottom: 16px;
  font-size: 13px;
}

.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 12px;
}

.table {
  margin-top: 20px;
}

@media (max-width: 720px) {
  .grid { grid-template-columns: 1fr; }
}
</style>
