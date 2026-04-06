<template>
  <div class="models-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.models') }}</span>
          <div class="header-actions">
            <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">{{ t('common.batchDelete') }} ({{ selectedIds.length }})</el-button>
            <el-button type="primary" @click="showModelDialog()">{{ t('models.create') }}</el-button>
          </div>
        </div>
      </template>

      <el-table :data="models" stripe v-loading="loading" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" :label="t('models.name')" width="220" />
        <el-table-column :label="t('models.mappingCount')" width="120">
          <template #default="{ row }">
            {{ row.mapping_count }}
          </template>
        </el-table-column>
        <el-table-column :label="t('models.capabilities')">
          <template #default="{ row }">
            <div v-if="row.mapping_count > 0" class="capability-tags">
              <el-tag v-if="row.supports_stream" type="primary" size="small" style="margin-right: 4px">Stream</el-tag>
              <el-tag v-if="row.supports_tools" type="warning" size="small" style="margin-right: 4px">Tools</el-tag>
              <el-tag v-if="row.supports_vision" type="success" size="small">Vision</el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('models.tokenSummary')">
          <template #default="{ row }">
            <el-tooltip v-if="row.min_context_window > 0 || row.min_max_output > 0" 
              :content="`${row.min_context_window.toLocaleString()} / ${row.min_max_output.toLocaleString()}`" 
              placement="top">
              <span>{{ formatContextDisplay(row.min_context_window, row.min_max_output) }}</span>
            </el-tooltip>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="120">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleModelEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="showModelDialog(row)">{{ t('common.edit') }}</el-button>
            <el-button link type="default" @click="goDetail(row.id)">{{ t('common.detail') }}</el-button>
            <el-button link type="danger" @click="handleDeleteModel(row.id)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && models.length === 0" :description="t('common.noData')" />
    </el-card>

    <el-dialog v-model="modelDialogVisible" :title="editingModel ? t('models.editModel') : t('models.createModel')" width="400px">
      <el-form :model="modelForm" :rules="modelRules" ref="modelFormRef" label-width="auto">
        <el-form-item :label="t('models.name')" prop="name">
          <el-input v-model="modelForm.name" placeholder="e.g., gpt-4" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-switch v-model="modelForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modelDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleModelSubmit" :loading="submitting">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api'
import { formatContextDisplay } from '@/utils/format'

const { t } = useI18n()
const router = useRouter()

interface Model {
  id: number
  name: string
  enabled: boolean
  mapping_count: number
  min_context_window: number
  min_max_output: number
  supports_vision: boolean
  supports_tools: boolean
  supports_stream: boolean
}

const models = ref<Model[]>([])
const selectedIds = ref<number[]>([])
const loading = ref(false)
const modelDialogVisible = ref(false)
const submitting = ref(false)
const editingModel = ref<Model | null>(null)
const modelFormRef = ref()

const modelForm = reactive({
  name: '',
  enabled: true
})

const modelRules = computed(() => ({
  name: [{ required: true, message: t('common.required'), trigger: 'blur' }]
}))

onMounted(() => {
  fetchModels()
})

async function fetchModels() {
  loading.value = true
  try {
    const res = await api.get('/models')
    models.value = (res.data.models || []).map((m: any) => ({
      id: m.id,
      name: m.model,
      enabled: m.enabled,
      mapping_count: m.mapping_count,
      min_context_window: m.min_context_window || 0,
      min_max_output: m.min_max_output || 0,
      supports_vision: m.supports_vision || false,
      supports_tools: m.supports_tools || false,
      supports_stream: m.supports_stream || false
    }))
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(selection: Model[]) {
  selectedIds.value = selection.map(item => item.id)
}

async function showModelDialog(model?: Model) {
  editingModel.value = model || null
  Object.assign(modelForm, {
    name: model?.name || '',
    enabled: model?.enabled ?? true
  })
  modelDialogVisible.value = true
}

async function handleModelSubmit() {
  const valid = await modelFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingModel.value) {
      await api.put(`/models/${editingModel.value.id}`, { name: modelForm.name, enabled: modelForm.enabled })
    } else {
      await api.post('/models', { name: modelForm.name, enabled: modelForm.enabled })
    }
    ElMessage.success(t('common.success'))
    modelDialogVisible.value = false
    fetchModels()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    submitting.value = false
  }
}

async function handleDeleteModel(id: number) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  await api.delete(`/models/${id}`)
  ElMessage.success(t('common.success'))
  fetchModels()
}

async function handleBatchDelete() {
  if (selectedIds.value.length === 0) return
  await ElMessageBox.confirm(t('common.confirm') + ` (${selectedIds.value.length} items)`, t('common.batchDelete'), { type: 'warning' })
  try {
    await Promise.all(selectedIds.value.map(id => api.delete(`/models/${id}`)))
    ElMessage.success(t('common.success'))
    selectedIds.value = []
    fetchModels()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function toggleModelEnabled(model: Model) {
  await api.put(`/models/${model.id}`, { enabled: model.enabled })
}

function goDetail(id: number) {
  router.push(`/models/${id}`)
}
</script>

<style scoped>
.models-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.capability-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
</style>