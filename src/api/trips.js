import request from './request'

export function listTrips() {
  return request.get('/api/trips')
}

export function createTrip(data) {
  return request.post('/api/trips', data)
}

export function updateTrip(id, data) {
  return request.put(`/api/trips/${id}`, data)
}

export function deleteTrip(id) {
  return request.delete(`/api/trips/${id}`)
}
