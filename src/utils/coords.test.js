import { describe, expect, it } from 'vitest'
import { gcj02ToWgs84, outOfChina, wgs84ToGcj02 } from './coords'

// 与 coordtransform 官方参照值一致（Go 版 geo_test.go 用同一组数，保证双语言算法一致）
describe('wgs84ToGcj02', () => {
  it('北京参照点', () => {
    const [lng, lat] = wgs84ToGcj02(116.404, 39.915)
    expect(lng).toBeCloseTo(116.41024449916938, 10)
    expect(lat).toBeCloseTo(39.91640428150164, 10)
  })

  it('上海参照点', () => {
    const [lng, lat] = wgs84ToGcj02(121.491, 31.242)
    expect(lng).toBeCloseTo(121.49546460985425, 9)
    expect(lat).toBeCloseTo(31.240013408324618, 9)
  })
})

describe('gcj02ToWgs84', () => {
  it('北京参照点', () => {
    const [lng, lat] = gcj02ToWgs84(116.404, 39.915)
    expect(lng).toBeCloseTo(116.39775550083061, 10)
    expect(lat).toBeCloseTo(39.91359571849836, 10)
  })

  it('与正向转换互为近似逆（往返误差 < 2e-5 度，约 1~2 米）', () => {
    const lng = 120.15507
    const lat = 30.274085
    const [glng, glat] = wgs84ToGcj02(lng, lat)
    const [blng, blat] = gcj02ToWgs84(glng, glat)
    expect(Math.abs(blng - lng)).toBeLessThan(2e-5)
    expect(Math.abs(blat - lat)).toBeLessThan(2e-5)
  })
})

describe('outOfChina', () => {
  it('境内点返回 false，境外点返回 true', () => {
    expect(outOfChina(116.404, 39.915)).toBe(false)
    expect(outOfChina(139.6917, 35.6895)).toBe(true) // 东京
    expect(outOfChina(-0.1278, 51.5074)).toBe(true) // 伦敦
  })

  it('境外坐标转换应原样返回', () => {
    expect(wgs84ToGcj02(139.6917, 35.6895)).toEqual([139.6917, 35.6895])
    expect(gcj02ToWgs84(139.6917, 35.6895)).toEqual([139.6917, 35.6895])
  })
})
