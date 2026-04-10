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

      <el-table :data="models" stripe v-loading="loading" @selection-change="handleSelectionChange" :default-sort="defaultSort" @sort-change="handleSortChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" :label="t('models.name')" width="220" sortable />
        <el-table-column :label="t('models.mappingCount')" width="120" prop="mapping_count" sortable>
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
        <el-table-column :label="t('models.tokenSummary')" prop="min_context_window" sortable>
          <template #default="{ row }">
            <el-tooltip v-if="row.min_context_window > 0 || row.min_max_output > 0" 
              :content="`${row.min_context_window.toLocaleString()} / ${row.min_max_output.toLocaleString()}`" 
              placement="top">
              <span>{{ formatContextDisplay(row.min_context_window, row.min_max_output) }}</span>
            </el-tooltip>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="120" prop="enabled" sortable>
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleModelEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="240">
          <template #default="{ row }">
            <el-button link type="success" @click="testModel(row)">{{ t('provider.test') }}</el-button>
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

    <el-dialog v-model="testDialogVisible" :title="t('provider.testModel') + ': ' + (testModelInfo?.name || '')" width="700px">
      <div v-if="testing" style="text-align: center; padding: 20px;">
        <el-icon class="is-loading" :size="32"><Loading /></el-icon>
        <p>{{ t('common.loading') }}</p>
      </div>
      <div v-else>
        <div v-if="testResults.length === 0" style="text-align: center; padding: 20px; color: #909399;">
          {{ t('common.noData') }}
        </div>
        <div v-for="mapping in testResults" :key="mapping.mapping_id" class="test-mapping-item">
          <div class="test-mapping-header">
            <span class="test-mapping-provider">{{ mapping.provider?.name }}</span>
            <span class="test-mapping-arrow">→</span>
            <span class="test-mapping-model">{{ mapping.provider_model?.display_name || mapping.provider_model?.model_id }}</span>
          </div>
          <div class="test-protocol-results">
            <div v-for="(result, idx) in mapping.protocol_tests" :key="idx" class="test-protocol-item">
              <div class="test-protocol-header">
                <el-tag :type="result.success ? 'success' : 'danger'" size="small">
                  {{ result.protocol.toUpperCase() }}
                </el-tag>
                <span class="test-protocol-status">
                  {{ result.success ? t('common.success') : t('common.error') }}
                  <span v-if="result.call_method === 'convert'" class="test-convert-badge">({{ t('provider.protocolConvert') }})</span>
                </span>
                <span class="test-protocol-latency">{{ result.latency_ms }}ms</span>
                <span class="test-protocol-tokens">{{ result.input_tokens }}/{{ result.output_tokens }}</span>
              </div>
              <div v-if="result.response" class="test-response">{{ result.response }}</div>
              <div v-if="result.error" class="test-error">{{ result.error }}</div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="testDialogVisible = false">{{ t('common.cancel') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import api from '@/api'
import { formatContextDisplay } from '@/utils/format'
import { getSortConfig, setSortConfig } from '@/utils/tableSort'

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
const defaultSort = getSortConfig('models', 'name')

const testDialogVisible = ref(false)
const testing = ref(false)
const testModelInfo = ref<Model | null>(null)
const testResults = ref<any[]>([])

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

async function testModel(model: Model) {
  testModelInfo.value = model
  testResults.value = []
  testing.value = true
  testDialogVisible.value = true
  
  try {
    const res = await api.post(`/models/${model.id}/test`)
    testResults.value = res.data.tests || []
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
    testDialogVisible.value = false
  } finally {
    testing.value = false
  }
}

function handleSortChange({ prop, order }: any) {
  if (prop && order) {
    setSortConfig('models', { prop, order })
  }
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

.test-mapping-item {
  margin-bottom: 16px;
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}

.test-mapping-item:last-child {
  margin-bottom: 0;
}

.test-mapping-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-weight: 500;
}

.test-mapping-provider {
  color: #409eff;
}

.test-mapping-arrow {
  color: #909399;
}

.test-mapping-model {
  color: #67c23a;
}

.test-protocol-results {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.test-protocol-item {
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.test-protocol-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.test-protocol-status {
  color: #606266;
}

.test-convert-badge {
  color: #909399;
  font-size: 12px;
}

.test-protocol-latency {
  color: #909399;
  margin-left: auto;
}

.test-protocol-tokens {
  color: #909399;
}

.test-response {
  margin-top: 8px;
  padding: 8px;
  background: #fff;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.5;
}

.test-error {
  margin-top: 8px;
  padding: 8px;
  background: #fef0f0;
  border-radius: 4px;
  color: #f56c6c;
  font-size: 13px;
}
</style>