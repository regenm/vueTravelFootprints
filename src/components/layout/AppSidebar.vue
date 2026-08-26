<template>
  <aside class="sidebar" :class="{ open: ui.sidebarOpen }">
    <div class="stats">
      <div>
        <b>{{ markers.stats.places }}</b>
        <span>地点</span>
      </div>
      <div>
        <b>{{ markers.stats.photos }}</b>
        <span>照片</span>
      </div>
      <div>
        <b>{{ markers.stats.categories }}</b>
        <span>分类</span>
      </div>
    </div>

    <div v-if="auth.isLoggedIn && !isSharePage" class="tabs">
      <button :class="{ on: tab === 'mine' }" type="button" @click="showMine">我的足迹</button>
      <button :class="{ on: tab === 'shared' }" type="button" @click="showShared">共享地图</button>
    </div>

    <div v-if="isSharePage && share" class="share-banner">
      <div class="share-head">
        <UserAvatar :src="share.owner?.avatar" :name="share.owner?.displayName" :size="36" />
        <div>
          <p>{{ share.title }}</p>
          <small>
            {{ share.owner?.displayName || '旅人' }}
            · {{ markers.canEdit ? '可一起记录' : '只读' }}
          </small>
        </div>
      </div>
      <p v-if="share.description" class="desc">{{ share.description }}</p>
      <div v-if="participants.length" class="faces">
        <UserAvatar
          v-for="p in participants.slice(0, 6)"
          :key="p.id"
          :src="p.avatar"
          :name="p.displayName"
          :size="24"
        />
        <small v-if="participants.length > 6">+{{ participants.length - 6 }}</small>
      </div>
      <el-button v-if="auth.isLoggedIn" size="small" plain @click="router.push('/')">返回我的地图</el-button>
    </div>

    <div v-if="tab === 'shared' && !isSharePage" class="inbox">
      <h4>我创建的</h4>
      <button
        v-for="item in markers.myShares"
        :key="item.id"
        class="inbox-item"
        type="button"
        @click="openShared(item)"
      >
        <UserAvatar :src="item.owner?.avatar" :name="item.owner?.displayName || item.title" :size="36" />
        <span class="meta">
          <strong>{{ item.title }}</strong>
          <small>{{ item.markerCount }} 个地点 · {{ item.isPublic ? '公开链接' : '仅成员' }}</small>
        </span>
      </button>
      <p v-if="!markers.myShares.length" class="empty slim">还没有创建共享地图</p>

      <h4>朋友分享的</h4>
      <button
        v-for="item in markers.inboxShares"
        :key="item.id"
        class="inbox-item"
        type="button"
        @click="openShared(item)"
      >
        <UserAvatar :src="item.owner?.avatar" :name="item.owner?.displayName" :size="36" />
        <span class="meta">
          <strong>{{ item.title }}</strong>
          <small>{{ item.owner?.displayName }} · {{ item.markerCount }} 个地点</small>
        </span>
      </button>
      <p v-if="!markers.inboxShares.length" class="empty slim">还没有人把旅行地图分享给你</p>
    </div>

    <div v-else class="timeline">
      <p v-if="!groups.length" class="empty">还没有足迹。点击「记录足迹」，再在地图上选一个点。</p>
      <section v-for="[year, list] in groups" :key="year">
        <h4>{{ year }}</h4>
        <button
          v-for="m in list"
          :key="m.id"
          type="button"
          class="item"
          :class="{ active: markers.selectedId === m.id }"
          @click="markers.selectMarker(m)"
        >
          <span class="thumb" :style="{ background: meta(m).color }">
            <img v-if="m.photos?.[0]" :src="resolveImageUrl(m.photos[0])" alt="" />
            <i v-else>{{ meta(m).emoji }}</i>
          </span>
          <span class="meta">
            <strong>{{ m.name }}</strong>
            <small>
              {{ monthDayOf(m) || '未注明日期' }} · {{ meta(m).label }}
              <template v-if="isSharePage && m.author"> · {{ m.author.displayName }}</template>
            </small>
          </span>
          <UserAvatar
            v-if="isSharePage"
            :src="m.author?.avatar"
            :name="m.author?.displayName"
            :size="22"
          />
        </button>
      </section>
    </div>
  </aside>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import UserAvatar from '@/components/common/UserAvatar.vue'
import { useAuthStore } from '@/stores/auth'
import { useMarkersStore } from '@/stores/markers'
import { useUiStore } from '@/stores/ui'
import { groupMarkersByYear, monthDayOf } from '@/utils/format'
import { resolveImageUrl } from '@/utils/image'
import { getCategoryMeta } from '@/utils/constants'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const markers = useMarkersStore()
const ui = useUiStore()
const tab = ref('mine')

const isSharePage = computed(() => Boolean(route.params.token))
const groups = computed(() => groupMarkersByYear(markers.filteredMarkers))
const share = computed(() => markers.currentShare?.share || null)
const participants = computed(() => share.value?.participants || [])

function meta(m) {
  return getCategoryMeta(m.category)
}

async function showMine() {
  tab.value = 'mine'
  if (markers.source !== 'mine') await markers.fetchMarkers()
}

async function showShared() {
  tab.value = 'shared'
  await markers.refreshShares()
}

function openShared(item) {
  tab.value = 'mine'
  router.push(`/s/${item.token}`)
}

watch(
  () => route.params.token,
  (token) => {
    if (token) tab.value = 'mine'
  }
)

onMounted(() => {
  if (auth.isLoggedIn) markers.refreshShares().catch(() => {})
})
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-w);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--tf-card);
  border-right: 1px solid var(--tf-line);
  z-index: 12;
  min-height: 0;
}

.sidebar:not(.open) {
  display: none;
}

.stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  padding: 16px;
}

.stats div {
  background: var(--tf-paper);
  border-radius: 14px;
  padding: 10px 8px;
  text-align: center;
}

.stats b {
  display: block;
  font-family: var(--font-serif);
  font-size: 22px;
}

.stats span {
  color: var(--tf-ink-faint);
  font-size: 12px;
}

.tabs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  margin: 0 16px 8px;
  padding: 4px;
  background: var(--tf-paper);
  border-radius: 12px;
}

.tabs button {
  border: none;
  background: transparent;
  border-radius: 9px;
  height: 32px;
  cursor: pointer;
  color: var(--tf-ink-soft);
}

.tabs button.on {
  background: #fff;
  color: var(--tf-teal-deep);
  font-weight: 600;
}

.share-banner {
  margin: 0 16px 8px;
  padding: 12px;
  border-radius: 12px;
  background: var(--tf-teal-soft);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.share-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.share-banner p {
  font-weight: 600;
}

.share-banner small,
.share-banner .desc {
  color: var(--tf-ink-soft);
  font-weight: 400;
}

.share-banner .desc {
  font-size: 12px;
  line-height: 1.5;
}

.faces {
  display: flex;
  align-items: center;
  gap: 4px;
}

.timeline,
.inbox {
  flex: 1;
  overflow: auto;
  padding: 8px 12px 20px;
}

.timeline h4,
.inbox h4 {
  font-family: var(--font-serif);
  margin: 12px 8px 8px;
  color: var(--tf-ink-soft);
}

.item,
.inbox-item {
  width: 100%;
  display: flex;
  gap: 10px;
  align-items: center;
  text-align: left;
  border: none;
  background: transparent;
  border-radius: 14px;
  padding: 8px;
  cursor: pointer;
}

.item:hover,
.inbox-item:hover,
.item.active {
  background: var(--tf-paper);
}

.thumb {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  overflow: hidden;
  display: grid;
  place-items: center;
  color: #fff;
  flex-shrink: 0;
}

.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.meta strong,
.inbox-item strong {
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta small,
.inbox-item span {
  color: var(--tf-ink-faint);
  font-size: 12px;
}

.empty {
  color: var(--tf-ink-faint);
  font-size: 13px;
  padding: 20px 10px;
  line-height: 1.7;
}

.empty.slim {
  padding: 4px 10px 16px;
}

@media (max-width: 860px) {
  .sidebar {
    position: absolute;
    inset: 0 auto 0 0;
    height: 100%;
    box-shadow: var(--tf-shadow);
  }
}
</style>
