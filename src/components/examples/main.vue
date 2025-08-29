<template>
    <div id="map-container"></div>
  </template>
  
  <script>
  import { onMounted, onBeforeUnmount, ref } from "vue";
  import AMapLoader from "@amap/amap-jsapi-loader";
  
  export default {
    setup() {
      const map = ref(null);
  
      const initMap = () => {
        AMapLoader.load({
          key: "", // 替换为你的高德地图 API Key
          version: "2.0",
        })
          .then((AMap) => {
            map.value = new AMap.Map("map-container", {
              mapStyle: "amap://styles/fresh",
              zoom: 11,
              center: [121.896964, 30.882957], // 初始中心点
            });
  
            // 可选：添加比例尺控件
            map.value.addControl(new AMap.Scale());
  
            // 可选：地图点击事件
            map.value.on("click", (e) => {
              console.log("点击位置经纬度：", e.lnglat);
            });
          })
          .catch((err) => console.error("地图加载失败:", err));
      };
  
      onMounted(() => {
        initMap();
      });
  
      onBeforeUnmount(() => {
        map.value?.destroy();
      });
  
      return {};
    },
  };
  </script>
  
  <style scoped>
  #map-container {
    width: 100%;
    height: 100vh;
    border: 1px solid #ccc;
    border-radius: 8px;
  }
  </style>
  