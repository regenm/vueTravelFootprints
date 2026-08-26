<template>
  <div class="auth">
    <section class="hero">
      <div class="hero-inner">
        <span class="mark">迹</span>
        <h1>把旅途，写进地图</h1>
        <p>记录去过的地方、收藏路上的照片，也可以把同一张旅行地图分享给同行的人。</p>
        <ul>
          <li>搜索地点或点击地图记下足迹</li>
          <li>照片轮播，重温当时的风景</li>
          <li>生成链接，和朋友一起看、一起写</li>
        </ul>
      </div>
      <svg class="route" viewBox="0 0 640 360" aria-hidden="true">
        <path d="M40 260 C 140 80, 220 300, 320 160 S 520 40, 600 180" fill="none" stroke="rgba(255,253,248,.35)" stroke-width="3" stroke-dasharray="8 10" />
        <circle cx="40" cy="260" r="8" fill="#e8c36a" />
        <circle cx="320" cy="160" r="8" fill="#fff" />
        <circle cx="600" cy="180" r="8" fill="#e8c36a" />
      </svg>
    </section>

    <section class="panel">
      <h2>欢迎回来</h2>
      <p class="sub">使用管理员开通的账号登录。不能自行注册。</p>

      <el-form :model="form" @submit.prevent="submit">
        <el-form-item label="用户名或邮箱">
          <el-input v-model="form.account" placeholder="请输入账号" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" @keyup.enter="submit" />
        </el-form-item>
        <el-button type="primary" class="submit" :loading="loading" @click="submit">进入地图</el-button>
      </el-form>

      <p class="demo">没有账号请联系管理员开通。</p>
    </section>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const loading = ref(false)
const form = reactive({
  account: '',
  password: ''
})

async function submit() {
  loading.value = true
  try {
    await auth.login(form.account, form.password)
    ElMessage.success('欢迎回来')
    router.replace(route.query.redirect || '/')
  } catch (err) {
    ElMessage.error(err.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth {
  height: 100%;
  display: grid;
  grid-template-columns: 1.15fr 0.85fr;
  background: var(--tf-paper);
}

.hero {
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(circle at 20% 20%, rgba(232, 195, 106, 0.28), transparent 32%),
    linear-gradient(160deg, #0b5553 0%, #0f6e6b 48%, #163a3a 100%);
  color: #fffdf8;
  padding: 64px;
  display: flex;
  align-items: flex-end;
}

.hero-inner { max-width: 460px; position: relative; z-index: 1; }

.mark {
  width: 52px;
  height: 52px;
  display: grid;
  place-items: center;
  border-radius: 16px;
  background: rgba(255, 253, 248, 0.12);
  font-family: var(--font-serif);
  font-size: 28px;
  margin-bottom: 24px;
}

h1 {
  font-family: var(--font-serif);
  font-size: 48px;
  line-height: 1.15;
  margin-bottom: 16px;
}

.hero p {
  opacity: 0.86;
  font-size: 16px;
  margin-bottom: 24px;
}

.hero li {
  list-style: none;
  margin: 8px 0;
  opacity: 0.9;
}

.hero li::before {
  content: '· ';
  color: #e8c36a;
}

.route {
  position: absolute;
  right: -40px;
  top: 40px;
  width: 72%;
  opacity: 0.9;
}

.panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 48px 56px;
  background: var(--tf-card);
}

h2 {
  font-family: var(--font-serif);
  font-size: 32px;
}

.sub {
  color: var(--tf-ink-soft);
  margin: 8px 0 28px;
}

.submit { width: 100%; height: 42px; margin-top: 8px; }

.demo {
  margin-top: 18px;
  color: var(--tf-ink-faint);
  font-size: 13px;
}

@media (max-width: 860px) {
  .auth { grid-template-columns: 1fr; }
  .hero { min-height: 240px; padding: 32px 24px; }
  h1 { font-size: 32px; }
  .panel { padding: 28px 20px 40px; }
  .route { display: none; }
}
</style>
