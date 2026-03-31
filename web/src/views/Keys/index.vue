<template>
  <div class="api-keys-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.keys') }}</span>
          <div class="header-actions">
            <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">{{ t('common.batchDelete') }} ({{ selectedIds.length }})</el-button>
            <el-button type="primary" @click="showDialog()">{{ t('apiKey.createKey') }}</el-button>
          </div>
        </div>
      </template>
      <el-table :data="keys" stripe v-loading="loading" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" :label="t('apiKey.name')" width="180" />
        <el-table-column prop="key" :label="t('apiKey.key')" width="240" />
        <el-table-column :label="t('apiKey.allowedModels')">
          <template #default="{ row }">
            <template v-if="row.models && row.models.length > 0">
              <el-tag v-for="m in row.models.slice(0, 3)" :key="m.id" size="small" style="margin-right: 4px">
                {{ m.model_alias }}
              </el-tag>
              <el-tag v-if="row.models.length > 3" size="small" type="info">+{{ row.models.length - 3 }}</el-tag>
            </template>
            <span v-else style="color: #999">{{ t('apiKey.allModels') }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="150">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="150">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDialog(row)">{{ t('common.edit') }}</el-button>
            <el-button link type="danger" @click="handleDelete(row.id)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? t('common.edit') : t('apiKey.createKey')">
      <el-form :model="form" ref="formRef" label-width="auto">
        <el-form-item :label="t('apiKey.name')">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item :label="t('apiKey.allowedModels')">
          <el-select v-model="form.allowed_models" multiple style="width: 100%" :placeholder="t('apiKey.allModels')" filterable>
            <el-option v-for="m in availableModels" :key="m.alias" :label="m.alias" :value="m.alias" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('apiKey.quota')">
          <el-input-number v-model="form.quota" :min="0" />
        </el-form-item>
        <el-form-item :label="t('apiKey.rateLimit')">
          <el-input-number v-model="form.rate_limit" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="keyDialogVisible" title="API Key">
      <p>{{ t('apiKey.key') }}:</p>
      <el-input v-model="newKey" readonly>
        <template #append>
          <el-button @click="copyKey">Copy</el-button>
        </template>
      </el-input>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api'
import { formatLatency, formatTokens } from '@/utils/format'

const { t } = useI18n()

const keys = ref<any[]>([])
const availableModels = ref<any[]>([])
const selectedIds = ref<number[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const keyDialogVisible = ref(false)
const newKey = ref('')
const editingId = ref<number | null>(null)
const formRef = ref()

const form = reactive({
  name: '',
  quota: 0,
  rate_limit: 0,
  allowed_models: [] as string[]
})

onMounted(() => {
  fetchKeys()
  fetchAvailableModels()
})

async function fetchKeys() {
  loading.value = true
  try {
    const res = await api.get('/api-keys')
    keys.value = res.data.keys || []
  } finally {
    loading.value = false
  }
}

async function fetchAvailableModels() {
  try {
    const res = await api.get('/model-mappings')
    const seen = new Set<string>()
    availableModels.value = (res.data.mappings || [])
      .filter((m: any) => {
        if (seen.has(m.alias)) return false
        seen.add(m.alias)
        return true
      })
      .map((m: any) => ({ alias: m.alias }))
  } catch (e) {
    console.error(e)
  }
}

function handleSelectionChange(selection: any[]) {
  selectedIds.value = selection.map(item => item.id)
}

function showDialog(key?: any) {
  editingId.value = key?.id || null
  if (key) {
    Object.assign(form, {
      name: key.name || '',
      quota: key.quota || 0,
      rate_limit: key.rate_limit || 0,
      allowed_models: key.models?.map((m: any) => m.model_alias) || []
    })
  } else {
    Object.assign(form, { name: '', quota: 0, rate_limit: 0, allowed_models: [] })
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  try {
    if (editingId.value) {
      await api.put(`/api-keys/${editingId.value}`, form)
      dialogVisible.value = false
    } else {
      const res = await api.post('/api-keys', form)
      newKey.value = res.data.raw_key
      dialogVisible.value = false
      keyDialogVisible.value = true
    }
    fetchKeys()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function toggleEnabled(row: any) {
  await api.put(`/api-keys/${row.id}`, { enabled: row.enabled })
}

function copyKey() {
  navigator.clipboard.writeText(newKey.value)
  ElMessage.success('Copied!')
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  await api.delete(`/api-keys/${id}`)
  ElMessage.success(t('common.success'))
  fetchKeys()
}

async function handleBatchDelete() {
  if (selectedIds.value.length === 0) return
  await ElMessageBox.confirm(t('common.confirm') + ` (${selectedIds.value.length} items)`, t('common.batchDelete'), { type: 'warning' })
  try {
    await Promise.all(selectedIds.value.map(id => api.delete(`/api-keys/${id}`)))
    ElMessage.success(t('common.success'))
    selectedIds.value = []
    fetchKeys()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}
</script>

<style scoped>
.api-keys-page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; gap: 10px; }
</style>
