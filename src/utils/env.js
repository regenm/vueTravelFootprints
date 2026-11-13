export function apiBaseURL() {
  const v = import.meta.env.VITE_API_BASE_URL
  // 未配置时走相对路径：单进程同源部署下任意域名/端口均可访问
  if (v === undefined) return ''
  return String(v).replace(/\/$/, '')
}
