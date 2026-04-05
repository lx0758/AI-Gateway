<template>
  <div class="provider-detail">
    <el-page-header @back="$router.back()" :title="t('menu.providers')">
      <template #content>
        {{ provider?.name || t('menu.providers') }}
      </template>
    </el-page-header>

    <el-card class="info-card">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="t('provider.apiStyles')">
          <el-tag v-if="provider?.openai_base_url" type="success" style="margin-right: 4px">OpenAI</el-tag>
          <el-tag v-if="provider?.anthropic_base_url" type="primary">Anthropic</el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">
          <el-tag :type="provider?.enabled ? 'success' : 'info'">
            {{ provider?.enabled ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('provider.lastSync')">{{ formatDateTime(provider?.last_sync_at) }}</el-descriptions-item>
      </el-descriptions>
      <div class="actions">
        <el-button type="primary" @click="syncModels" :loading="syncing">{{ t('provider.syncModels') }}</el-button>
        <el-button type="success" @click="showAddDialog">{{ t('provider.addModel') }}</el-button>
        <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">{{ t('common.batchDelete') }} ({{ selectedIds.length }})</el-button>
      </div>
    </el-card>

    <el-card v-loading="loading">
      <template #header>{{ t('provider.models') }}</template>
      <el-table :data="models" stripe @row-click="showModelDetail" @selection-change="handleSelectionChange" class="clickable-table">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="model_id" :label="t('provider.modelId')" width="180" />
        <el-table-column prop="display_name" :label="t('common.name')" />
        <el-table-column :label="t('provider.capabilities')" width="200">
          <template #default="{ row }">
            <div class="capability-tags">
              <el-tag v-if="row.supports_stream" type="primary" size="small" style="margin-right: 4px">Stream</el-tag>
              <el-tag v-if="row.supports_tools" type="warning" size="small" style="margin-right: 4px">Tools</el-tag>
              <el-tag v-if="row.supports_vision" type="success" size="small">Vision</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('provider.contextWindow')" width="150" >
          <template #default="{ row }">
            <el-tooltip v-if="row.context_window > 0 || row.max_output > 0" 
              :content="`${row.context_window.toLocaleString()} / ${row.max_output.toLocaleString()}`" 
              placement="top">
              <span>{{ formatContextDisplay(row.context_window, row.max_output) }}</span>
            </el-tooltip>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('provider.price')" width="100" >
          <template #default="{ row }">
            {{ row.input_price }} / {{ row.output_price }}
          </template>
        </el-table-column>
        <el-table-column :label="t('provider.source')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.source === 'manual' ? 'warning' : 'info'" size="small">
              {{ row.source === 'manual' ? 'Manual' : 'Sync' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_available ? 'success' : 'danger'" size="small">
              {{ row.is_available ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="150">
          <template #default="{ row }">
            <el-button link type="primary" @click.stop="showEditDialog(row)">{{ t('common.edit') }}</el-button>
            <el-button link type="danger" @click.stop="handleDelete(row)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingModel ? t('common.edit') : t('provider.addModel')" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="auto">
        <el-form-item :label="t('provider.modelId')" prop="model_id">
          <el-input v-model="form.model_id" :disabled="!!editingModel" />
        </el-form-item>
        <el-form-item :label="t('common.name')">
          <el-input v-model="form.display_name" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="t('provider.contextWindow')">
              <el-input-number v-model="form.context_window" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('provider.maxOutput')">
              <el-input-number v-model="form.max_output" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="t('provider.inputPrice')">
              <el-input-number v-model="form.input_price" :min="0" :precision="6" :step="0.0001" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('provider.outputPrice')">
              <el-input-number v-model="form.output_price" :min="0" :precision="6" :step="0.0001" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item :label="t('provider.supportsVision')">
              <el-switch v-model="form.supports_vision" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('provider.supportsTools')">
              <el-switch v-model="form.supports_tools" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('provider.supportsStream')">
              <el-switch v-model="form.supports_stream" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" :title="detailModel?.model_id" width="600px">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="t('provider.modelId')">{{ detailModel?.model_id }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.name')">{{ detailModel?.display_name || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="t('provider.contextWindow')">{{ detailModel?.context_window?.toLocaleString() || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="t('provider.maxOutput')">{{ detailModel?.max_output?.toLocaleString() || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="t('provider.inputPrice')">${{ (detailModel?.input_price || 0).toFixed(4) }} / 1K tokens</el-descriptions-item>
        <el-descriptions-item :label="t('provider.outputPrice')">${{ (detailModel?.output_price || 0).toFixed(4) }} / 1K tokens</el-descriptions-item>
        <el-descriptions-item :label="t('provider.source')">
          <el-tag :type="detailModel?.source === 'manual' ? 'warning' : 'info'" size="small">
            {{ detailModel?.source === 'manual' ? 'Manual' : 'Sync' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">
          <el-tag :type="detailModel?.is_available ? 'success' : 'danger'" size="small">
            {{ detailModel?.is_available ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('provider.supportsVision')">
          <el-tag :type="detailModel?.supports_vision ? 'success' : 'info'" size="small">
            {{ detailModel?.supports_vision ? t('common.yes') : t('common.no') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('provider.supportsTools')">
          <el-tag :type="detailModel?.supports_tools ? 'success' : 'info'" size="small">
            {{ detailModel?.supports_tools ? t('common.yes') : t('common.no') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('provider.supportsStream')">
          <el-tag :type="detailModel?.supports_stream ? 'success' : 'info'" size="small">
            {{ detailModel?.supports_stream ? t('common.yes') : t('common.no') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="'Owned By'">{{ detailModel?.owned_by || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.createdAt')">{{ formatDateTime(detailModel?.created_at) }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.updatedAt')">{{ formatDateTime(detailModel?.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <div style="margin-top: 16px; display: flex; gap: 8px;">
        <el-button type="primary" @click="showEditDialog(detailModel); detailDialogVisible = false">{{ t('common.edit') }}</el-button>
        <el-button type="danger" @click="handleDelete(detailModel); detailDialogVisible = false">{{ t('common.delete') }}</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api'
import { formatDateTime, formatContextDisplay } from '@/utils/format'

const { t } = useI18n()
const route = useRoute()

const provider = ref<any>(null)
const models = ref<any[]>([])
const selectedIds = ref<number[]>([])
const loading = ref(false)
const syncing = ref(false)
const dialogVisible = ref(false)
const detailDialogVisible = ref(false)
const detailModel = ref<any>(null)
const submitting = ref(false)
const editingModel = ref<any>(null)
const formRef = ref()

const providerId = route.params.id as string

const form = reactive({
  model_id: '',
  display_name: '',
  context_window: 0,
  max_output: 0,
  input_price: 0,
  output_price: 0,
  supports_vision: false,
  supports_tools: true,
  supports_stream: true
})

const rules = {
  model_id: [{ required: true, message: () => t('common.required'), trigger: 'blur' }]
}

onMounted(() => {
  fetchProvider()
})

async function fetchProvider() {
  loading.value = true
  try {
    const res = await api.get(`/providers/${providerId}`)
    provider.value = res.data.provider
    models.value = (res.data.provider?.models || []).sort((a: any, b: any) => a.model_id.localeCompare(b.model_id))
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(selection: any[]) {
  selectedIds.value = selection.map(item => item.id)
}

async function syncModels() {
  syncing.value = true
  try {
    await api.post(`/providers/${providerId}/sync`)
    ElMessage.success(t('common.success'))
    fetchProvider()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    syncing.value = false
  }
}

function showAddDialog() {
  editingModel.value = null
  Object.assign(form, {
    model_id: '',
    display_name: '',
    context_window: 0,
    max_output: 0,
    input_price: 0,
    output_price: 0,
    supports_vision: false,
    supports_tools: true,
    supports_stream: true
  })
  dialogVisible.value = true
}

function showEditDialog(model: any) {
  editingModel.value = model
  Object.assign(form, {
    model_id: model.model_id,
    display_name: model.display_name || '',
    context_window: model.context_window || 0,
    max_output: model.max_output || 0,
    input_price: model.input_price || 0,
    output_price: model.output_price || 0,
    supports_vision: model.supports_vision || false,
    supports_tools: model.supports_tools || false,
    supports_stream: model.supports_stream !== false
  })
  dialogVisible.value = true
}

function showModelDetail(model: any) {
  detailModel.value = model
  detailDialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingModel.value) {
      await api.put(`/providers/${providerId}/models/${editingModel.value.id}`, form)
    } else {
      await api.post(`/providers/${providerId}/models`, form)
    }
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    fetchProvider()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    submitting.value = false
  }
}

async function handleDelete(model: any) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  try {
    await api.delete(`/providers/${providerId}/models/${model.id}`)
    ElMessage.success(t('common.success'))
    fetchProvider()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function handleBatchDelete() {
  if (selectedIds.value.length === 0) return
  await ElMessageBox.confirm(t('common.confirm') + ` (${selectedIds.value.length} items)`, t('common.batchDelete'), { type: 'warning' })
  try {
    await Promise.all(selectedIds.value.map(id => api.delete(`/providers/${providerId}/models/${id}`)))
    ElMessage.success(t('common.success'))
    selectedIds.value = []
    fetchProvider()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}
</script>

<style scoped>
.provider-detail {
  padding: 20px;
}

.info-card {
  margin: 20px 0;
}

.actions {
  margin-top: 20px;
  display: flex;
  gap: 10px;
}

.clickable-table :deep(.el-table__row) {
  cursor: pointer;
}

.clickable-table :deep(.el-table__row:hover) {
  background-color: #f5f7fa;
}

.capability-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
</style>
