import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import * as constants from '@/utils/constants'
import { useMarkersStore } from './markers'

// 省份/城市/行程总数常量（侧栏统计「x/34」「x/333」的数据来源）
describe('省市总数常量', () => {
  it('省份总数 34、城市总数 333', () => {
    expect(constants.PROVINCE_TOTAL).toBe(34)
    expect(constants.CITY_TOTAL).toBe(333)
  })
})

describe('stats 省市行程计数', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('按 provinceCode/cityCode/tripId 去重计数，空串不计入', () => {
    const store = useMarkersStore()
    store.markers = [
      {
        name: '西湖',
        category: '自然风光',
        photos: ['a.jpg'],
        provinceCode: '330000',
        cityCode: '330100',
        tripId: 't1'
      },
      {
        name: '故宫',
        category: '历史古迹',
        photos: [],
        provinceCode: '110000',
        cityCode: '110100',
        tripId: 't1'
      },
      {
        // 早期旧数据：无省市码与行程归属，不应计入省市/行程统计
        name: '未归类足迹',
        category: '自然风光',
        photos: ['b.jpg', 'c.jpg'],
        provinceCode: '',
        cityCode: '',
        tripId: ''
      },
      {
        // 与西湖同省不同市，省份应去重、城市应区分
        name: '千岛湖',
        category: '自然风光',
        photos: [],
        provinceCode: '330000',
        cityCode: '330200',
        tripId: 't2'
      }
    ]

    expect(store.stats).toEqual({
      places: 4,
      photos: 3,
      categories: 2,
      provinces: 2,
      cities: 3,
      trips: 2
    })
  })

  it('无足迹时各项为 0', () => {
    const store = useMarkersStore()
    expect(store.stats).toEqual({
      places: 0,
      photos: 0,
      categories: 0,
      provinces: 0,
      cities: 0,
      trips: 0
    })
  })
})
