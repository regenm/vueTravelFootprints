import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import * as authApi from '@/api/auth'
import { TOKEN_KEY, USER_KEY } from '@/utils/constants'

function readUser() {
  try {
    return JSON.parse(localStorage.getItem(USER_KEY) || 'null')
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem(TOKEN_KEY) || '')
  const user = ref(readUser())

  const isLoggedIn = computed(() => Boolean(token.value && user.value))
  const isAdmin = computed(() => user.value?.role === 'admin')
  const displayName = computed(() => user.value?.displayName || user.value?.username || '')

  function persist() {
    if (token.value) localStorage.setItem(TOKEN_KEY, token.value)
    else localStorage.removeItem(TOKEN_KEY)
    if (user.value) localStorage.setItem(USER_KEY, JSON.stringify(user.value))
    else localStorage.removeItem(USER_KEY)
  }

  function setSession(payload) {
    token.value = payload.token
    user.value = payload.user
    persist()
  }

  async function login(account, password) {
    const res = await authApi.login({ account, password })
    setSession(res.data)
    return res.data
  }

  async function hydrate() {
    if (!token.value) return null
    const res = await authApi.fetchMe()
    user.value = res.data
    persist()
    return user.value
  }

  async function updateProfile(payload) {
    const res = await authApi.updateMe(payload)
    user.value = res.data
    persist()
    return user.value
  }

  function logout() {
    token.value = ''
    user.value = null
    persist()
  }

  return {
    token,
    user,
    isLoggedIn,
    isAdmin,
    displayName,
    login,
    hydrate,
    updateProfile,
    logout
  }
})
