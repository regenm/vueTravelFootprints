import AMapLoader from '@amap/amap-jsapi-loader'

let loadPromise = null

export function applyAmapSecurity() {
  const code = String(import.meta.env.VITE_AMAP_SECURITY_CODE || '').trim()
  if (code) {
    window._AMapSecurityConfig = { securityJsCode: code }
  }
}

export function loadAMap() {
  if (window.AMap?.PlaceSearch && window.AMap?.Geocoder) {
    return Promise.resolve(window.AMap)
  }
  if (loadPromise) return loadPromise

  applyAmapSecurity()
  loadPromise = AMapLoader.load({
    key: import.meta.env.VITE_AMAP_KEY,
    version: '2.0',
    plugins: ['AMap.Scale', 'AMap.Geocoder', 'AMap.AutoComplete', 'AMap.PlaceSearch']
  })
    .then(
      (AMap) =>
        new Promise((resolve) => {
          window.AMap = AMap
          AMap.plugin(['AMap.Geocoder', 'AMap.AutoComplete', 'AMap.PlaceSearch'], () => resolve(AMap))
        })
    )
    .catch((err) => {
      loadPromise = null
      throw err
    })

  return loadPromise
}

function toSuggestion(name, location, addressParts = []) {
  const loc = extractLngLat(location)
  if (!name || !loc) return null
  const address = addressParts.filter(Boolean).join('')
  return {
    value: address ? `${name} · ${address}` : name,
    name,
    address,
    lng: loc.lng,
    lat: loc.lat
  }
}

export function extractLngLat(loc) {
  if (!loc || Array.isArray(loc)) return null
  if (typeof loc === 'string') {
    const parts = loc.split(',').map((n) => Number(String(n).trim()))
    if (parts.length >= 2 && Number.isFinite(parts[0]) && Number.isFinite(parts[1])) {
      return { lng: parts[0], lat: parts[1] }
    }
    return null
  }
  const lng = typeof loc.getLng === 'function' ? loc.getLng() : loc.lng
  const lat = typeof loc.getLat === 'function' ? loc.getLat() : loc.lat
  if (Number.isFinite(Number(lng)) && Number.isFinite(Number(lat))) {
    return { lng: Number(lng), lat: Number(lat) }
  }
  return null
}

function oncePromise(work, ms = 5000) {
  return new Promise((resolve) => {
    let settled = false
    const finish = (value) => {
      if (settled) return
      settled = true
      resolve(value)
    }
    const timer = setTimeout(() => finish([]), ms)
    Promise.resolve()
      .then(work)
      .then((list) => {
        clearTimeout(timer)
        finish(list || [])
      })
      .catch(() => {
        clearTimeout(timer)
        finish([])
      })
  })
}

export async function searchByAmap(keyword) {
  const AMap = await loadAMap()
  return oncePromise(
    () =>
      new Promise((resolve, reject) => {
        const run = () => {
          try {
            const search = new AMap.PlaceSearch({
              pageSize: 12,
              pageIndex: 1,
              citylimit: false
            })
            search.search(keyword, (status, result) => {
              const pois = result?.poiList?.pois || result?.pois || []
              const mapped = (status === 'complete' ? pois : [])
                .map((poi) => toSuggestion(poi.name, poi.location, [poi.pname, poi.cityname, poi.adname, poi.address]))
                .filter(Boolean)
              if (mapped.length || !AMap.AutoComplete) {
                resolve(mapped)
                return
              }
              const auto = new AMap.AutoComplete({})
              auto.search(keyword, (autoStatus, autoResult) => {
                const tips = autoStatus === 'complete' ? autoResult?.tips || [] : []
                resolve(
                  tips
                    .map((tip) => toSuggestion(tip.name, tip.location, [tip.district, tip.address]))
                    .filter(Boolean)
                )
              })
            })
          } catch (err) {
            reject(err)
          }
        }
        if (AMap.PlaceSearch) run()
        else AMap.plugin('AMap.PlaceSearch', run)
      })
  )
}
