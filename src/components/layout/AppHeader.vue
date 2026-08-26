<template>
  <header class="header">
    <div class="left">
      <button class="icon-btn" type="button" @click="ui.sidebarOpen = !ui.sidebarOpen" title="足迹列表">
        <el-icon><Menu /></el-icon>
      </button>
      <div class="brand" @click="goHome">
        <span class="logo">迹</span>
        <div class="brand-text">
          <strong>旅迹</strong>
          <small>{{ subtitle }}</small>
        </div>
      </div>
    </div>

    <div class="center">
      <el-input
        v-model="ui.keyword"
        placeholder="搜索地点、笔记、地址"
        clearable
        class="search"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select v-model="ui.category" clearable placeholder="全部分类" class="cat">
        <el-option
          v-for="cat in markers.categoryOptions"
          :key="cat.value"
          :label="`${cat.emoji} ${cat.label}`"
          :value="cat.value"
        />
      </el-select>
    </div>

    <div class="right">
      <el-button
        v-if="markers.canEdit"
        :type="ui.addMode ? 'warning' : 'primary'"
        :plain="ui.addMode"
        @click="ui.addMode ? (ui.addMode = false) : ui.openForm()"
      >
        <el-icon><Plus /></el-icon>
        {{ ui.addMode ? '取消地图选点' : '记录足迹' }}
      </el-button>
      <el-button v-if="auth.isLoggedIn && !markers.isShareView" plain @click="ui.openShare(null)">
        <el-icon><Share /></el-icon>
        分享
      </el-button>

      <template v-if="auth.isLoggedIn">
        <el-dropdown trigger="click">
          <button class="avatar-btn" type="button" title="账号">
            <UserAvatar :src="auth.user?.avatar" :name="auth.displayName" :size="38" />
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item disabled>{{ auth.displayName }}</el-dropdown-item>
              <el-dropdown-item @click="ui.profileOpen = true">个人资料</el-dropdown-item>
              <el-dropdown-item @click="ui.openShare(null)">共享地图</el-dropdown-item>
              <el-dropdown-item v-if="auth.isAdmin" @click="ui.usersOpen = true">用户管理</el-dropdown-item>
              <el-dropdown-item divided @click="onLogout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
      <el-button v-else type="primary" plain @click="router.push('/login')">登录</el-button>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Menu, Plus, Search, Share } from '@element-plus/icons-vue'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { useAuthStore } from '@/stores/auth'
import { useMarkersStore } from '@/stores/markers'
import { useUiStore } from '@/stores/ui'

const router = useRouter()
const auth = useAuthStore()
const markers = useMarkersStore()
const ui = useUiStore()

const subtitle = computed(() => {
  if (markers.currentShare?.share?.title) return markers.currentShare.share.title
  return '把旅途写进地图'
})

function goHome() {
  if (markers.isShareView) router.push('/')
}

function onLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.header {
  height: var(--header-h);
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 16px 0 12px;
  background: rgba(255, 253, 248, 0.92);
  border-bottom: 1px solid var(--tf-line);
  backdrop-filter: blur(16px);
  z-index: 20;
}

.left,
.right,
.center {
  display: flex;
  align-items: center;
  gap: 10px;
}

.left { min-width: 210px; }
.center { flex: 1; justify-content: center; }
.right { justify-content: flex-end; }

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.logo {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  background: var(--tf-teal);
  color: #fff;
  display: grid;
  place-items: center;
  font-family: var(--font-serif);
  font-weight: 700;
}

.brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}

.brand-text strong {
  font-family: var(--font-serif);
  font-size: 18px;
}

.brand-text small {
  color: var(--tf-ink-faint);
  font-size: 11px;
}

.search { width: min(360px, 42vw); }
.cat { width: 140px; }

.icon-btn,
.avatar-btn {
  width: 38px;
  height: 38px;
  border: none;
  border-radius: 50%;
  background: var(--tf-paper);
  color: var(--tf-ink);
  cursor: pointer;
  padding: 0;
}

.avatar-btn {
  background: transparent;
  display: grid;
  place-items: center;
}

@media (max-width: 860px) {
  .header { gap: 8px; padding: 0 8px; }
  .left { min-width: auto; }
  .brand-text { display: none; }
  .cat { display: none; }
  .search { width: 42vw; }
  .right :deep(.el-button span) { display: none; }
}
</style>
