import request from './request'

export function login(data) {
  return request.post('/api/auth/login', data)
}

export function fetchMe() {
  return request.get('/api/auth/me')
}

export function updateMe(data) {
  return request.put('/api/auth/me', data)
}

export function changePassword(data) {
  return request.put('/api/auth/password', data)
}

export function listUsers() {
  return request.get('/api/admin/users')
}

export function createUser(data) {
  return request.post('/api/admin/users', data)
}

export async function exportBackup() {
  const blob = await request.get('/api/admin/export', { responseType: 'blob', timeout: 0 })
  const d = new Date()
  const p = (n) => String(n).padStart(2, '0')
  const filename = `travel-backup-${d.getFullYear()}${p(d.getMonth() + 1)}${p(d.getDate())}-${p(d.getHours())}${p(d.getMinutes())}${p(d.getSeconds())}.zip`
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}
