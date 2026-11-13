<template>
  <el-dialog
    :model-value="ui.profileOpen"
    title="个人资料"
    width="440px"
    append-to-body
    destroy-on-close
    @close="ui.profileOpen = false"
  >
    <div class="profile">
      <label class="avatar-edit" title="更换头像">
        <UserAvatar :src="preview || auth.user?.avatar" :name="form.displayName || auth.displayName" :size="96" />
        <span class="camera">更换</span>
        <input type="file" accept="image/*" hidden @change="onFile" />
      </label>
      <p class="hint">点击头像上传照片，将显示在地图标记和共享地图里。</p>

      <el-form label-position="top" @submit.prevent="save">
        <el-form-item label="用户名">
          <el-input :model-value="auth.user?.username" disabled />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="form.displayName" maxlength="24" placeholder="别人看到的名字" />
        </el-form-item>
      </el-form>

      <el-divider class="pwd-divider">
        <button class="pwd-toggle" type="button" @click="pwdOpen = !pwdOpen">
          {{ pwdOpen ? '收起修改密码' : '修改密码' }}
        </button>
      </el-divider>

      <el-form v-if="pwdOpen" label-position="top" @submit.prevent="changePwd">
        <el-form-item label="旧密码">
          <el-input v-model="pwd.old" type="password" show-password autocomplete="current-password" />
        </el-form-item>
        <el-form-item label="新密码（至少 8 位）">
          <el-input v-model="pwd.next" type="password" show-password autocomplete="new-password" />
        </el-form-item>
        <el-form-item label="确认新密码">
          <el-input v-model="pwd.confirm" type="password" show-password autocomplete="new-password" @keyup.enter="changePwd" />
        </el-form-item>
        <el-button class="pwd-btn" type="warning" plain :loading="changingPwd" @click="changePwd">确认修改密码</el-button>
      </el-form>
    </div>

    <template #footer>
      <el-button v-if="auth.user?.avatar || preview" plain @click="clearAvatar">移除头像</el-button>
      <el-button @click="ui.profileOpen = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import * as markersApi from '@/api/markers'
import { changePassword } from '@/api/auth'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { compressAvatar } from '@/utils/image'

const auth = useAuthStore()
const ui = useUiStore()
const saving = ref(false)
const preview = ref('')
const pendingAvatar = ref(null)
const form = reactive({
  displayName: ''
})
const pwdOpen = ref(false)
const changingPwd = ref(false)
const pwd = reactive({ old: '', next: '', confirm: '' })

watch(
  () => ui.profileOpen,
  (open) => {
    if (!open) return
    form.displayName = auth.displayName
    preview.value = ''
    pendingAvatar.value = undefined
    pwdOpen.value = false
    pwd.old = ''
    pwd.next = ''
    pwd.confirm = ''
  }
)

async function onFile(e) {
  const file = e.target.files?.[0]
  e.target.value = ''
  if (!file) return
  saving.value = true
  try {
    const compressed = await compressAvatar(file)
    const res = await markersApi.uploadImage(compressed)
    pendingAvatar.value = res.url
    preview.value = res.url
    ElMessage.success('头像已上传，点击保存生效')
  } catch (err) {
    ElMessage.error(err.message || '头像上传失败')
  } finally {
    saving.value = false
  }
}

function clearAvatar() {
  pendingAvatar.value = ''
  preview.value = ''
}

async function save() {
  saving.value = true
  try {
    const payload = { displayName: form.displayName.trim() }
    if (pendingAvatar.value !== undefined) payload.avatar = pendingAvatar.value
    await auth.updateProfile(payload)
    ElMessage.success('资料已更新')
    ui.profileOpen = false
  } catch (err) {
    ElMessage.error(err.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function changePwd() {
  if (!pwd.old || !pwd.next) {
    ElMessage.warning('请填写旧密码和新密码')
    return
  }
  if (pwd.next !== pwd.confirm) {
    ElMessage.warning('两次输入的新密码不一致')
    return
  }
  changingPwd.value = true
  try {
    await changePassword({ oldPassword: pwd.old, newPassword: pwd.next })
    ElMessage.success('密码已修改')
    pwdOpen.value = false
    pwd.old = ''
    pwd.next = ''
    pwd.confirm = ''
  } catch (err) {
    ElMessage.error(err.message || '修改失败')
  } finally {
    changingPwd.value = false
  }
}
</script>

<style scoped>
.profile {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.avatar-edit {
  position: relative;
  cursor: pointer;
  border-radius: 50%;
}

.avatar-edit:hover .camera,
.avatar-edit:focus-within .camera {
  opacity: 1;
}

.camera {
  position: absolute;
  inset: auto 0 0;
  height: 28px;
  display: grid;
  place-items: center;
  background: rgba(28, 25, 21, 0.62);
  color: #fff;
  font-size: 12px;
  border-radius: 0 0 48px 48px;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.hint {
  margin: 12px 0 18px;
  color: var(--tf-ink-faint);
  font-size: 12px;
  text-align: center;
}

.profile :deep(.el-form) {
  width: 100%;
}

.pwd-divider {
  margin: 8px 0 16px;
}

.pwd-divider :deep(.el-divider__text) {
  padding: 0;
}

.pwd-toggle {
  border: none;
  background: none;
  color: var(--tf-teal);
  cursor: pointer;
  font-size: 13px;
  padding: 4px 8px;
}

.pwd-toggle:hover {
  text-decoration: underline;
}

.pwd-btn {
  width: 100%;
}
</style>
