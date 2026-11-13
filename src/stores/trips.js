import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as tripsApi from '@/api/trips'

export const useTripsStore = defineStore('trips', () => {
  const trips = ref([])
  const loading = ref(false)

  async function fetchTrips() {
    loading.value = true
    try {
      const res = await tripsApi.listTrips()
      trips.value = res.data || []
      return trips.value
    } finally {
      loading.value = false
    }
  }

  async function createTrip(data) {
    const res = await tripsApi.createTrip(data)
    await fetchTrips()
    return res.data
  }

  async function updateTrip(id, data) {
    await tripsApi.updateTrip(id, data)
    await fetchTrips()
  }

  async function removeTrip(id) {
    await tripsApi.deleteTrip(id)
    trips.value = trips.value.filter((t) => t.id !== id)
  }

  return { trips, loading, fetchTrips, createTrip, updateTrip, removeTrip }
})
