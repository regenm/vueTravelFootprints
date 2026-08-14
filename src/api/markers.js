import request from './request'

export function getMarkers() {
  return request.get('/api/markers')
}

export function getMarkerById(id) {
  return request.get(`/api/markers/${id}`)
}

export function createMarker(data) {
  return request.post('/api/markers', data)
}

export function updateMarker(id, data) {
  return request.put(`/api/markers/${id}`, data)
}

export function deleteMarker(id) {
  return request.delete(`/api/markers/${id}`)
}

export function searchMarkers(params) {
  return request.get('/api/markers/search', { params })
}

export function uploadImage(file) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/api/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}