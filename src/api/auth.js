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

export function listUsers() {
  return request.get('/api/admin/users')
}

export function createUser(data) {
  return request.post('/api/admin/users', data)
}
