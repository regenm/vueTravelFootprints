<template>
    <div>
      <h2>关于页面</h2>
      <p>这里是示例页面，可以扩展其他功能</p>
    </div>
  </template>
  
  <script>
  export default {
    setup() {
    const amapKey = import.meta.env.VITE_AMAAP_KEY;
    console.log(amapKey);
    console.log( import.meta.env)
    } 
  };
  </script>
  