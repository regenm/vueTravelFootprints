<template>
  <div id="map-container">
    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner">加载中...</div>
    </div>
    <div v-if="error" class="error-overlay">
      <div class="error-message">
        {{ error }}
        <button @click="fetchMarkersData">重试</button>
      </div>
    </div>
  </div>
</template>

<script>
import { onMounted, onBeforeUnmount, ref, reactive } from "vue";
import { createApp } from "vue";
import AMapLoader from "@amap/amap-jsapi-loader";
import MarkerInfo from "./MarkerInfo.vue";
import MarkerAvatar from "./MarkerAvatar.vue";

export default {
  setup() {
    const amapKey = import.meta.env.VITE_AMAP_KEY;
    const map = ref(null);
    const loading = ref(false);
    const error = ref(null);
    const markers = reactive([]);

    // 从后端获取标记点数据
    const fetchMarkersData = async () => {
      try {
        loading.value = true;
        error.value = null;
        const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:5000";

        const response = await fetch(API_BASE + "/markers");
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const resJson = await response.json();
        if (!resJson.success) {
          throw new Error(resJson.message || "接口返回失败");
        }

        const data = resJson.data || [];

        // 清空并更新本地 markers
        markers.splice(0, markers.length, ...data);

        // 如果地图已经初始化，更新地图标记
        if (map.value) {
          updateMapMarkers(data);
        }

      } catch (err) {
        console.error("获取标记数据失败:", err);
        error.value = `获取数据失败: ${err.message}`;
      } finally {
        loading.value = false;
      }
    };

    // 更新地图标记
    const updateMapMarkers = (markersData) => {
      map.value.clearMap();

      markersData.forEach((marker) => {
        const lng = Number(marker.longitude);
        const lat = Number(marker.latitude);
        if (isNaN(lng) || isNaN(lat)) return;

        // Marker 图标用头像组件
        const container = document.createElement("div");
        const markerApp = createApp(MarkerAvatar, {
          photo: marker.photos && marker.photos.length > 0 ? marker.photos[0] : ""
        });
        markerApp.mount(container);

        const amapMarker = new window.AMap.Marker({
          position: [lng, lat],
          map: map.value,
          content: container,
          offset: new window.AMap.Pixel(-20, -20),
          extData: marker
        });

        amapMarker.on("click", () => {
          const infoContainer = document.createElement("div");
          const infoApp = createApp(MarkerInfo, {
            name: marker.name,
            photos: marker.photos || []
          });
          infoApp.mount(infoContainer);

          const infoWindow = new window.AMap.InfoWindow({
            content: infoContainer,
            offset: new window.AMap.Pixel(0, -30),
            closeWhenClickMap: true
          });

          infoWindow.open(map.value, amapMarker.getPosition());
          infoWindow.on("close", () => infoApp.unmount());
        });
      });
    };

    const initMap = async () => {
      try {
        const AMap = await AMapLoader.load({
          key: amapKey,
          version: "2.0",
          plugins: ["AMap.Scale", "AMap.Marker", "AMap.InfoWindow"],
        });

        window.AMap = AMap;

        map.value = new AMap.Map("map-container", {
          zoom: 13,
          center: [120.15507, 30.27415],
          mapStyle: "amap://styles/fresh",
        });

        map.value.addControl(new AMap.Scale());

        await fetchMarkersData();

      } catch (err) {
        console.error("地图初始化失败:", err);
        error.value = `地图加载失败: ${err.message}`;
      }
    };

    onMounted(() => {
      initMap();
    });

    onBeforeUnmount(() => {
      if (map.value) {
        map.value.destroy();
      }
    });

    return {
      loading,
      error,
      fetchMarkersData
    };
  },
};
</script>


<style scoped>
#map-container {
  width: 100%;
  height: 100vh;
  position: relative;
}

.loading-overlay,
.error-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.8);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.loading-spinner {
  font-size: 18px;
  color: #333;
}

.error-message {
  text-align: center;
  color: #d32f2f;
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.error-message button {
  margin-top: 10px;
  padding: 8px 16px;
  background: #1976d2;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.error-message button:hover {
  background: #1565c0;
}
</style>