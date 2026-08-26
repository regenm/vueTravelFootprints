export const CATEGORIES = [
  { label: '自然风光', value: '自然风光', emoji: '🏔️', color: '#2f9e44' },
  { label: '历史古迹', value: '历史古迹', emoji: '🏛️', color: '#d9480f' },
  { label: '美食探店', value: '美食探店', emoji: '🍜', color: '#c92a2a' },
  { label: '城市漫步', value: '城市漫步', emoji: '🏙️', color: '#1971c2' },
  { label: '海滩度假', value: '海滩度假', emoji: '🏖️', color: '#0c8599' },
  { label: '文化体验', value: '文化体验', emoji: '⛩️', color: '#7048e8' },
  { label: '自驾路书', value: '自驾路书', emoji: '🚗', color: '#c2255c' }
]

export const CATEGORY_MAP = Object.fromEntries(CATEGORIES.map((c) => [c.value, c]))

export function getCategoryMeta(value) {
  return CATEGORY_MAP[value] || { label: value || '未分类', emoji: '📍', color: '#0f6e6b' }
}

export const TOKEN_KEY = 'tf_token'
export const USER_KEY = 'tf_user'
