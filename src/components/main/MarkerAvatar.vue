<template>
  <div class="marker-avatar">
    <img :src="photo" alt="marker" />
  </div>
</template>

<script>
export default {
  name: "MarkerAvatar",
  props: {
    photo: String,
  },
};
</script>

<style scoped>
.marker-avatar {
  width: 40px;
  height: 40px;
  overflow: hidden;
  border: 2px solid #fff;
  border-radius: 8px;
  box-shadow: 0 0 6px rgba(0, 0, 0, 0.3);
  background: #fff;
}
.marker-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>
