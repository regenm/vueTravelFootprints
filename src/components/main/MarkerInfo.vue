<template>
  <div class="info-card">
    <h3>{{ name }}</h3>
    <div class="photos">
      <div 
        v-for="(p, i) in photos" 
        :key="i" 
        class="image-item"
        @click="handleImageClick(i)"
      >
        <img
          :src="getImageUrl(p)"
          :alt="`${name} image ${i + 1}`"
          @error="handleImageError"
          class="thumbnail"
        />
      </div>
    </div>

    <el-image-viewer
      v-if="previewVisible"
      :url-list="previewList"
      :initial-index="previewIndex"
      @close="previewVisible = false"
      teleported
    />
  </div>
</template>

<script>
import { ElImageViewer } from 'element-plus'

export default {
  name: "MarkerInfo",
  components: { ElImageViewer },
  props: {
    name: String,
    photos: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      previewVisible: false,
      previewIndex: 0
    }
  },
  computed: {
    previewList() {
      return this.photos.map(p => this.getImageUrl(p));
    }
  },
  methods: {
    getImageUrl(path) {
      if (!path) return '';
      
      // 处理各种路径情况
      if (path.startsWith('http') || path.startsWith('data:')) {
        return path;
      }
      
      // 处理相对路径
      try {
        return require(`@/assets/${path}`);
      } catch (e) {
        console.warn('Image not found:', path);
        return path;
      }
    },
    
    handleImageError(e) {
      e.target.src = 'https://via.placeholder.com/90x60?text=Image+Error';
      e.target.onerror = null; // 防止循环错误
    },
    
    handleImageClick(index) {
      this.previewIndex = index;
      this.previewVisible = true;
    }
  }
}
</script>

<style scoped>
.info-card {
  width: 400px;
  padding: 15px;
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  text-align: center;
  z-index: 1000;
}

.info-card h3 {
  margin: 0 0 12px 0;
  color: #333;
  font-size: 16px;
}

.photos {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
}

.image-item {
  cursor: pointer;
  transition: transform 0.2s;
}

.image-item:hover {
  transform: scale(1.05);
}

.thumbnail {
  width: 90px;
  height: 60px;
  border-radius: 4px;
  object-fit: cover;
  border: 1px solid #eee;
}
</style>