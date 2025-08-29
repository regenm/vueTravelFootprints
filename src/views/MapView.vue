<template>
    <div>
      <HeaderBar  />
      <MapPart />
    </div>
  </template>
  
  <script>
  import HeaderBar from "../components/main/HeaderBar.vue";
  import MapPart from "../components/main/MapPart.vue";
  import { ref } from "vue";    
  
  export default {
    components: { HeaderBar, MapPart },
    setup() {
    },
  };
  </script>
  