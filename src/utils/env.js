export function apiBaseURL() {
  const v = import.meta.env.VITE_API_BASE_URL
  if (v === undefined) return 'http://localhost:5000'
  return String(v).replace(/\/$/, '')
}
