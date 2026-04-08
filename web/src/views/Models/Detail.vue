<template>
  <div class="model-detail">
    <el-page-header @back="$router.back()" :title="t('menu.models')">
      <template #content>
        {{ model?.model || t('menu.models') }}
      </template>
    </el-page-header>

    <el-card class="info-card">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="t('models.name')">{{ model?.model || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">
          <el-tag :type="modelEnabled ? 'success' : 'info'" size="small">
            {{ modelEnabled ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
      <div class="actions">
        <el-button type="success" @click="showMappingDialog()">{{ t('models.addMapping') }}</el-button>
        <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">{{ t('common.batchDelete') }} ({{ selectedIds.length }})</el-button>
      </div>
    </el-card>

    <el-card v-loading="loading">
      <template #header>{{ t('models.mappings') }}</template>
      <el-table 
        :data="mappings" 
        stripe 
        @selection-change="handleSelectionChange"
        row-key="id"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column width="50" align="center">
          <template #default>
            <el-icon class="drag-handle"><Rank /></el-icon>
          </template>
        </el-table-column>
        <el-table-column :label="t('provider.name')" width="180">
          <template #default="{ row }">{{ row.provider?.name }}</template>
        </el-table-column>
        <el-table-column :label="t('models.providerType')" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.provider?.openai_base_url" type="success" size="small" style="margin-right: 4px">OpenAI</el-tag>
            <el-tag v-if="row.provider?.anthropic_base_url" type="primary" size="small">Anthropic</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="provider_model_name" :label="t('modelMapping.actualModel')" />
        <el-table-column :label="t('models.capabilities')">
          <template #default="{ row }">
            <div v-if="row.model_info" class="capability-tags">
              <el-tag v-if="row.model_info.supports_stream" type="primary" size="small" style="margin-right: 4px">Stream</el-tag>
              <el-tag v-if="row.model_info.supports_tools" type="warning" size="small" style="margin-right: 4px">Tools</el-tag>
              <el-tag v-if="row.model_info.supports_vision" type="success" size="small">Vision</el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('models.contextWindow')">
          <template #default="{ row }">
            <el-tooltip v-if="row.model_info && (row.model_info.context_window > 0 || row.model_info.max_output > 0)" 
              :content="`${row.model_info.context_window.toLocaleString()} / ${row.model_info.max_output.toLocaleString()}`" 
              placement="top">
              <span>{{ formatContextDisplay(row.model_info.context_window, row.model_info.max_output) }}</span>
            </el-tooltip>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="weight" :label="t('modelMapping.weight')" width="120" />
        <el-table-column :label="t('common.status')" width="120">
          <template #default="{ row }">
            <el-tooltip v-if="!modelEnabled" :content="t('models.modelDisabled')" placement="top">
              <el-switch :model-value="false" disabled />
            </el-tooltip>
            <el-switch v-else v-model="row.enabled" @change="toggleMappingEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="120">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showMappingDialog(row)">{{ t('common.edit') }}</el-button>
            <el-button link type="danger" size="small" @click="handleDeleteMapping(row.id)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="mappingDialogVisible" :title="editingMapping ? t('models.editMapping') : t('models.addMapping')" width="500px">
      <el-form :model="mappingForm" :rules="mappingRules" ref="mappingFormRef" label-width="auto" v-loading="mappingDialogLoading">
        <el-form-item :label="t('provider.name')" prop="provider_id">
          <el-select v-model="mappingForm.provider_id" @change="loadProviderModels" style="width: 100%" filterable>
            <el-option v-for="p in providers" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('modelMapping.model')" prop="provider_model_name">
          <el-select v-model="mappingForm.provider_model_name" style="width: 100%" filterable :placeholder="mappingForm.provider_id ? '' : t('provider.name')">
            <el-option v-for="m in providerModels" :key="m.model_id" :label="m.model_id" :value="m.model_id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('modelMapping.weight')">
          <el-input-number v-model="mappingForm.weight" :min="0" :max="100" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-switch v-model="mappingForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mappingDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleMappingSubmit" :loading="submitting">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Rank } from '@element-plus/icons-vue'
import Sortable from 'sortablejs'
import api from '@/api'
import { formatContextDisplay } from '@/utils/format'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

interface ModelInfo {
  context_window: number
  max_output: number
  supports_vision: boolean
  supports_tools: boolean
  supports_stream: boolean
}

interface Mapping {
  id: number
  provider_id: number
  provider_model_name: string
  weight: number
  enabled: boolean
  provider?: {
    id: number
    name: string
    openai_base_url: string
    anthropic_base_url: string
  }
  model_info?: ModelInfo
}

interface Model {
  id: number
  model: string
  enabled: boolean
}

const model = ref<Model | null>(null)
const modelEnabled = ref(true)
const mappings = ref<Mapping[]>([])
const providers = ref<any[]>([])
const providerModels = ref<any[]>([])
const selectedIds = ref<number[]>([])
const loading = ref(false)
const mappingDialogVisible = ref(false)
const mappingDialogLoading = ref(false)
const submitting = ref(false)
const editingMapping = ref<Mapping | null>(null)
const mappingFormRef = ref()
let providersLoaded = false
let sortableInstance: Sortable | null = null

const modelId = route.params.id as string

const mappingForm = reactive({
  provider_id: null as number | null,
  provider_model_name: '',
  weight: 0,
  enabled: true
})

const mappingRules = computed(() => ({
  provider_id: [{ required: true, message: t('common.required'), trigger: 'change' }],
  provider_model_name: [{ required: true, message: t('common.required'), trigger: 'change' }]
}))

onMounted(() => {
  fetchModel()
})

async function fetchModel() {
  loading.value = true
  try {
    const res = await api.get(`/models/${modelId}`)
    model.value = res.data.model
    modelEnabled.value = res.data.model.enabled
    mappings.value = res.data.model.mappings || []
    
    setTimeout(() => {
      initSortable()
    }, 100)
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function initSortable() {
  const el = document.querySelector('.el-table__body-wrapper tbody')
  if (!el || sortableInstance) return
  
  sortableInstance = new Sortable(el as HTMLElement, {
    handle: '.drag-handle',
    animation: 150,
    onEnd: async (evt) => {
      const { oldIndex, newIndex } = evt
      if (oldIndex === newIndex) return
      
      const newMappings = [...mappings.value]
      const [moved] = newMappings.splice(oldIndex!, 1)
      newMappings.splice(newIndex!, 0, moved)
      
      const order = newMappings.map(m => m.id)
      
      try {
        await api.put(`/models/${modelId}/mappings/order`, { order })
        mappings.value = newMappings
        mappings.value.forEach((m, i) => {
          m.weight = mappings.value.length - 1 - i
        })
        ElMessage.success(t('common.success'))
      } catch (e: any) {
        ElMessage.error(e.response?.data?.error || t('common.error'))
        fetchModel()
      }
    }
  })
}

async function fetchProviders() {
  const res = await api.get('/providers')
  providers.value = (res.data.providers || []).sort((a: any, b: any) => a.name.localeCompare(b.name))
}

async function loadProviderModels() {
  if (!mappingForm.provider_id) return
  const res = await api.get(`/providers/${mappingForm.provider_id}/models`)
  providerModels.value = (res.data.models || []).sort((a: any, b: any) => a.model_id.localeCompare(b.model_id))
}

function handleSelectionChange(selection: Mapping[]) {
  selectedIds.value = selection.map(item => item.id)
}

async function showMappingDialog(mapping?: Mapping) {
  if (!providersLoaded) {
    await fetchProviders()
    providersLoaded = true
  }
  editingMapping.value = mapping || null
  Object.assign(mappingForm, {
    provider_id: mapping?.provider_id || null,
    provider_model_name: mapping?.provider_model_name || '',
    weight: mapping?.weight || mappings.value.length,
    enabled: mapping?.enabled ?? true
  })
  providerModels.value = []
  if (mappingForm.provider_id) {
    loadProviderModels()
  }
  mappingDialogVisible.value = true
}

async function handleMappingSubmit() {
  const valid = await mappingFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingMapping.value) {
      await api.put(`/models/${modelId}/mappings/${editingMapping.value.id}`, mappingForm)
    } else {
      await api.post(`/models/${modelId}/mappings`, mappingForm)
    }
    ElMessage.success(t('common.success'))
    mappingDialogVisible.value = false
    fetchModel()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    submitting.value = false
  }
}

async function handleDeleteMapping(mappingId: number) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  await api.delete(`/models/${modelId}/mappings/${mappingId}`)
  ElMessage.success(t('common.success'))
  fetchModel()
}

async function handleBatchDelete() {
  if (selectedIds.value.length === 0) return
  await ElMessageBox.confirm(t('common.confirm') + ` (${selectedIds.value.length} items)`, t('common.batchDelete'), { type: 'warning' })
  try {
    await Promise.all(selectedIds.value.map(id => api.delete(`/models/${modelId}/mappings/${id}`)))
    ElMessage.success(t('common.success'))
    selectedIds.value = []
    fetchModel()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function toggleMappingEnabled(mapping: Mapping) {
  await api.put(`/models/${modelId}/mappings/${mapping.id}`, { enabled: mapping.enabled })
}
</script>

<style scoped>
.model-detail {
  padding: 20px;
}

.info-card {
  margin: 20px 0;
}

.info-card :deep(.el-descriptions__table) {
  width: 100%;
  table-layout: fixed;
}

.info-card :deep(.el-descriptions__cell) {
  width: 50%;
}

.actions {
  margin-top: 20px;
  display: flex;
  gap: 10px;
}

.drag-handle {
  cursor: move;
  color: #909399;
}

.drag-handle:hover {
  color: #409eff;
}

.capability-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
</style>