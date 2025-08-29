<template>
  <div class="header-bar">
    <el-button @click="showAddDialog = true" type="primary">增加足迹！</el-button>
    
    <!-- 添加足迹的对话框 -->
    <el-dialog
      v-model="showAddDialog"
      title="添加新足迹"
      width="600px"
      :before-close="handleClose"
    >
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item label="地点名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入地点名称" />
        </el-form-item>
        
        <el-form-item label="经度" prop="longitude">
          <el-input-number 
            v-model="form.longitude" 
            :min="-180" 
            :max="180" 
            :step="0.000001"
            :precision="6"
            placeholder="请输入经度"
          />
        </el-form-item>
        
        <el-form-item label="纬度" prop="latitude">
          <el-input-number 
            v-model="form.latitude" 
            :min="-90" 
            :max="90" 
            :step="0.000001"
            :precision="6"
            placeholder="请输入纬度"
          />
        </el-form-item>
        
        <el-form-item label="图片URL" prop="photos">
          <div class="photo-urls-container">
            <div 
              v-for="(url, index) in form.photos" 
              :key="index" 
              class="url-input-item"
            >
              <el-input 
                v-model="form.photos[index]" 
                placeholder="请输入图片URL" 
                class="url-input"
              >
                <template #append>
                  <el-button 
                    @click="removePhotoUrl(index)" 
                    :disabled="form.photos.length <= 1"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </template>
              </el-input>
            </div>
            
            <el-button 
              @click="addPhotoUrl" 
              type="primary" 
              plain 
              class="add-url-btn"
            >
              <el-icon><Plus /></el-icon>
              添加图片URL
            </el-button>
            
            <div class="url-count">
              已添加 {{ form.photos.length }} 个图片URL
            </div>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddDialog = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="submitForm" 
            :loading="loading"
            :disabled="!isFormValid"
          >
            {{ loading ? '提交中...' : '确认添加' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'

export default {
  name: "HeaderBar",
  components: {
    Delete,
    Plus
  },
  emits: ["marker-added"],
  
  setup(props, { emit }) {
    const showAddDialog = ref(false)
    const loading = ref(false)
    const formRef = ref(null)

    const form = reactive({
      name: '',
      longitude: null,
      latitude: null,
      photos: [''] // ✅ 和后端保持一致
    })

    const rules = {
      name: [
        { required: true, message: '请输入地点名称', trigger: 'blur' }
      ],
      longitude: [
        { required: true, message: '请输入经度', trigger: 'blur' }
      ],
      latitude: [
        { required: true, message: '请输入纬度', trigger: 'blur' }
      ]
    }

    const isFormValid = computed(() => {
      return form.name && form.longitude !== null && form.latitude !== null
    })

    const handleClose = (done) => {
      if (loading.value) return
      resetForm()
      done()
    }

    const resetForm = () => {
      form.name = ''
      form.longitude = null
      form.latitude = null
      form.photos = ['']
    }

    const addPhotoUrl = () => {
      form.photos.push('')
    }

    const removePhotoUrl = (index) => {
      if (form.photos.length > 1) {
        form.photos.splice(index, 1)
      }
    }

    const submitForm = async () => {
      if (!formRef.value) return
      try {
        const valid = await formRef.value.validate()
        if (!valid) return
        loading.value = true

        // 过滤空URL
        const filteredPhotos = form.photos.filter(url => url.trim() !== '')

        const markerData = {
          name: form.name,
          longitude: String(form.longitude), // ✅ 转成字符串，兼容后端 .strip()
          latitude: String(form.latitude),
          photos: filteredPhotos
        }

        console.log('提交的数据:', markerData)

        // ✅ 改为调用 Flask 后端接口
        const response = await fetch('http://127.0.0.1:5000/add-marker', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(markerData)
        })

        const result = await response.json()

        if (!result.success) {
          throw new Error(result.message || '添加失败')
        }

        emit('marker-added', result.data)
        ElMessage.success('足迹添加成功！')
        showAddDialog.value = false
        resetForm()
      } catch (error) {
        console.error('添加足迹错误:', error)
        ElMessage.error(`添加足迹失败: ${error.message}`)
      } finally {
        loading.value = false
      }
    }

    return {
      showAddDialog,
      loading,
      form,
      rules,
      formRef,
      isFormValid,
      handleClose,
      addPhotoUrl,
      removePhotoUrl,
      submitForm
    }
  }
}
</script>
<style scoped> .header-bar { display: flex; align-items: center; height: 50px; background-color: #0078d4; padding: 0 20px; color: white; } .header-bar button { margin-right: 10px; padding: 6px 12px; background-color: #005ea6; color: white; border: none; border-radius: 4px; cursor: pointer; } .header-bar button:hover { background-color: #004080; } .photo-urls-container { width: 100%; } .url-input-item { margin-bottom: 10px; } .url-input { margin-bottom: 8px; } .add-url-btn { margin-top: 10px; width: 100%; } .url-count { margin-top: 8px; font-size: 12px; color: #666; text-align: center; } :deep(.el-dialog__body) { padding: 20px; } :deep(.el-form-item) { margin-bottom: 20px; } :deep(.el-input-number) { width: 100%; } </style>