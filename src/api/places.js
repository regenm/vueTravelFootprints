import request from './request'

export function searchPlaces(q) {
  return request.get('/api/places', { params: { q } })
}
