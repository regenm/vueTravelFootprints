export const CATEGORIES = [
  { label: '自然风光', value: '自然风光', emoji: '🏔️', color: '#34C759' },
  { label: '历史古迹', value: '历史古迹', emoji: '🏛️', color: '#A2845E' },
  { label: '美食探店', value: '美食探店', emoji: '🍜', color: '#FF9500' },
  { label: '城市漫步', value: '城市漫步', emoji: '🏙️', color: '#5856D6' },
  { label: '海滩度假', value: '海滩度假', emoji: '🏖️', color: '#30B0C7' },
  { label: '文化体验', value: '文化体验', emoji: '⛩️', color: '#AF52DE' },
  { label: '自驾路书', value: '自驾路书', emoji: '🚗', color: '#FF3B30' },
  { label: '酒店民宿', value: '酒店民宿', emoji: '🏨', color: '#FFCC00' },
  { label: '购物血拼', value: '购物血拼', emoji: '🛍️', color: '#FF2D55' },
  { label: '户外徒步', value: '户外徒步', emoji: '⛺', color: '#00C7BE' },
  { label: '酒吧咖啡', value: '酒吧咖啡', emoji: '☕', color: '#8E8E93' },
  { label: '交通枢纽', value: '交通枢纽', emoji: '✈️', color: '#007AFF' }
]

export const CATEGORY_MAP = Object.fromEntries(CATEGORIES.map((c) => [c.value, c]))

export function getCategoryMeta(value) {
  return CATEGORY_MAP[value] || { label: value || '未分类', emoji: '📍', color: '#4877ad' }
}

export const TOKEN_KEY = 'tf_token'
export const USER_KEY = 'tf_user'

export const PROVINCE_TOTAL = 34
export const CITY_TOTAL = 333
