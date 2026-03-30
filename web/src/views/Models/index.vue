<template>
  <div class="models-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.models') }}</span>
          <el-button type="primary" @click="showDialog()">{{ t('common.create') }}</el-button>
        </div>
      </template>
      <el-table :data="mappings" stripe v-loading="loading">
        <el-table-column prop="alias" :label="t('modelMapping.alias')" />
        <el-table-column :label="t('provider.name')">
          <template #default="{ row }">{{ row.provider?.name }}</template>
        </el-table-column>
        <el-table-column :label="t('modelMapping.actualModel')">
          <template #default="{ row }">{{ row.provider_model?.model_id }}</template>
        </el-table-column>
        <el-table-column prop="weight" :label="t('modelMapping.weight')" />
        <el-table-column :label="t('common.status')">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="150">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDialog(row.id)">{{ t('common.edit') }}</el-button>
            <el-button link type="danger" @click="handleDelete(row.id)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? t('common.edit') : t('common.create')">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="auto" v-loading="dialogLoading">
        <el-form-item :label="t('modelMapping.alias')" prop="alias">
          <el-input v-model="form.alias" placeholder="e.g., gpt-4" />
        </el-form-item>
        <el-form-item :label="t('provider.name')" prop="provider_id">
          <el-select v-model="form.provider_id" @change="loadProviderModels" style="width: 100%" filterable>
            <el-option v-for="p in providers" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('modelMapping.model')" prop="provider_model_id">
          <el-select v-model="form.provider_model_id" style="width: 100%" filterable :placeholder="form.provider_id ? '' : t('provider.name')">
            <el-option v-for="m in providerModels" :key="m.id" :label="m.model_id" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('modelMapping.weight')">
          <el-input-number v-model="form.weight" :min="1" :max="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api'

const { t } = useI18n()

const mappings = ref<any[]>([])
const providers = ref<any[]>([])
const providerModels = ref<any[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogLoading = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref()

const form = reactive({
  alias: '',
  provider_id: null as number | null,
  provider_model_id: null as number | null,
  weight: 1
})

const rules = {
  alias: [{ required: true, message: () => t('modelMapping.required'), trigger: 'blur' }],
  provider_id: [{ required: true, message: () => t('modelMapping.required'), trigger: 'change' }],
  provider_model_id: [{ required: true, message: () => t('modelMapping.required'), trigger: 'change' }]
}

onMounted(() => {
  fetchMappings()
  fetchProviders()
})

async function fetchMappings() {
  loading.value = true
  try {
    const res = await api.get('/model-mappings')
    mappings.value = res.data.mappings || []
  } finally {
    loading.value = false
  }
}

async function fetchProviders() {
  const res = await api.get('/providers')
  providers.value = res.data.providers || []
}

async function loadProviderModels() {
  if (!form.provider_id) return
  const res = await api.get(`/providers/${form.provider_id}/models`)
  providerModels.value = res.data.models || []
}

async function showDialog(id?: number) {
  editingId.value = id || null
  Object.assign(form, { alias: '', provider_id: null, provider_model_id: null, weight: 1 })
  providerModels.value = []
  dialogVisible.value = true

  if (id) {
    dialogLoading.value = true
    try {
      const res = await api.get(`/model-mappings`)
      const mapping = res.data.mappings?.find((m: any) => m.id === id)
      if (mapping) {
        form.alias = mapping.alias
        form.provider_id = mapping.provider_id
        form.provider_model_id = mapping.provider_model_id
        form.weight = mapping.weight
        await loadProviderModels()
      }
    } catch (e) {
      ElMessage.error(t('common.error'))
      dialogVisible.value = false
    } finally {
      dialogLoading.value = false
    }
  }
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  try {
    if (editingId.value) {
      await api.put(`/model-mappings/${editingId.value}`, {
        alias: form.alias,
        provider_id: form.provider_id,
        provider_model_id: form.provider_model_id,
        weight: form.weight
      })
    } else {
      await api.post('/model-mappings', form)
    }
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    fetchMappings()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function toggleEnabled(row: any) {
  await api.put(`/model-mappings/${row.id}`, { enabled: row.enabled })
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  await api.delete(`/model-mappings/${id}`)
  ElMessage.success(t('common.success'))
  fetchMappings()
}
</script>

<style scoped>
.models-page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
