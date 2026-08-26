import request from './request'

export function createShare(data) {
  return request.post('/api/shares', data)
}

export function listMyShares() {
  return request.get('/api/shares')
}

export function listInboxShares() {
  return request.get('/api/shares/inbox')
}

export function deleteShare(id) {
  return request.delete(`/api/shares/${id}`)
}

export function updateShare(id, data) {
  return request.put(`/api/shares/${id}`, data)
}

export function leaveShare(id) {
  return request.delete(`/api/shares/${id}/members/me`)
}

export function addShareMember(id, data) {
  return request.post(`/api/shares/${id}/members`, data)
}

export function removeShareMember(shareId, userId) {
  return request.delete(`/api/shares/${shareId}/members/${userId}`)
}

export function getShareByToken(token) {
  return request.get(`/api/s/${token}`)
}
