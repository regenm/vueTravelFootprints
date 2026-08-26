import { searchPlaces as searchPlacesApi } from '@/api/places'
import { searchByAmap } from './amap'

export async function searchPlaces(keyword) {
  const q = String(keyword || '').trim()
  if (!q) return []

  try {
    const local = await searchByAmap(q)
    if (local.length) return local
  } catch (err) {
    console.warn('高德地点搜索失败，改用服务端搜索', err)
  }

  try {
    const res = await searchPlacesApi(q)
    return res.data || []
  } catch (err) {
    console.warn('服务端地点搜索失败', err)
    return []
  }
}
