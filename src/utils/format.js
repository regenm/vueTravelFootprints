export function formatVisitDate(value) {
  if (!value) return '未记录日期'
  return String(value).slice(0, 10)
}

export function yearOf(marker) {
  const d = marker?.visitDate || marker?.createdAt || ''
  const year = String(d).slice(0, 4)
  return /^\d{4}$/.test(year) ? year : '未注明年份'
}

export function monthDayOf(marker) {
  const d = formatVisitDate(marker?.visitDate || marker?.createdAt)
  if (d === '未记录日期' || d.length < 10) return ''
  return d.slice(5)
}

export function groupMarkersByYear(markers) {
  const map = new Map()
  markers.forEach((m) => {
    const year = yearOf(m)
    if (!map.has(year)) map.set(year, [])
    map.get(year).push(m)
  })
  return [...map.entries()].sort((a, b) => String(b[0]).localeCompare(String(a[0])))
}

export function initials(name = '') {
  const s = String(name).trim()
  if (!s) return '旅'
  return s.slice(0, 1).toUpperCase()
}

const AVATAR_TONES = ['#0f6e6b', '#c96a32', '#1971c2', '#7048e8', '#c2255c', '#2f9e44', '#0c8599', '#5c5348']

export function avatarTone(name = '') {
  const s = String(name)
  let hash = 0
  for (let i = 0; i < s.length; i += 1) hash = (hash + s.charCodeAt(i) * (i + 1)) % AVATAR_TONES.length
  return AVATAR_TONES[hash] || AVATAR_TONES[0]
}
