<template>
  <div class="providers">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.providers') }}</span>
          <div class="header-actions">
            <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">{{ t('common.batchDelete') }} ({{ selectedIds.length }})</el-button>
            <el-button type="primary" @click="showDialog()">{{ t('provider.addProvider') }}</el-button>
          </div>
        </div>
      </template>
      <el-table :data="providers" stripe v-loading="loading" @selection-change="handleSelectionChange" :default-sort="defaultSort" @sort-change="handleSortChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" :label="t('provider.name')" width="220" sortable />
        <el-table-column :label="t('provider.apiStyles')">
          <template #default="{ row }">
            <el-tag v-if="row.openai_base_url" type="success" style="margin-right: 4px">OpenAI</el-tag>
            <el-tag v-if="row.anthropic_base_url" type="primary">Anthropic</el-tag>
            <span v-if="!row.openai_base_url && !row.anthropic_base_url">-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('provider.models')" width="120" prop="models" sortable :sort-method="(a: any, b: any) => sortByArrayLength(a, b, 'models')">
          <template #default="{ row }">
            {{ row.models?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="120" prop="enabled" sortable>
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDialog(row.id)">{{ t('common.edit') }}</el-button>
            <el-button link type="default" @click="goDetail(row.id)">{{ t('common.detail') }}</el-button>
            <el-button link type="danger" @click="handleDelete(row.id)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? t('provider.editProvider') : t('provider.addProvider')">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="auto" v-loading="dialogLoading">
        <el-form-item :label="t('provider.name')" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item :label="'OpenAI BaseURL'">
          <el-input v-model="form.openai_base_url" placeholder="https://api.openai.com/v1" />
        </el-form-item>
        <el-form-item :label="'Anthropic BaseURL'">
          <el-input v-model="form.anthropic_base_url" placeholder="https://api.anthropic.com/v1" />
        </el-form-item>
        <el-form-item :label="t('provider.key')" prop="api_key">
          <el-input v-model="form.api_key" type="password" show-password :placeholder="editingId ? t('provider.keyPlaceholder') : ''" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ t('common.save') }}</el-button>
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
import { getSortConfig, setSortConfig, sortByArrayLength } from '@/utils/tableSort'

const { t } = useI18n()
const router = useRouter()

const providers = ref<any[]>([])
const selectedIds = ref<number[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogLoading = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref()
const defaultSort = getSortConfig('providers', 'name')

const form = reactive({
  name: '',
  openai_base_url: '',
  anthropic_base_url: '',
  api_key: ''
})

const rules = computed(() => ({
  name: [{ required: true, message: 'Required', trigger: 'blur' }],
  api_key: editingId.value ? [] : [{ required: true, message: 'Required', trigger: 'blur' }]
}))

function validateBaseURL() {
  if (!form.openai_base_url && !form.anthropic_base_url) {
    ElMessage.error('At least one BaseURL is required')
    return false
  }
  return true
}

onMounted(() => {
  fetchProviders()
})

async function fetchProviders() {
  loading.value = true
  try {
    const res = await api.get('/providers')
    providers.value = res.data.providers || []
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(selection: any[]) {
  selectedIds.value = selection.map(item => item.id)
}

async function showDialog(id?: number) {
  editingId.value = id || null
  Object.assign(form, { name: '', openai_base_url: '', anthropic_base_url: '', api_key: '' })
  dialogVisible.value = true
  
  if (id) {
    dialogLoading.value = true
    try {
      const res = await api.get(`/providers/${id}`)
      const provider = res.data.provider
      if (provider) {
        Object.assign(form, {
          name: provider.name || '',
          openai_base_url: provider.openai_base_url || '',
          anthropic_base_url: provider.anthropic_base_url || '',
          api_key: ''
        })
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
  
  if (!validateBaseURL()) return

  submitting.value = true
  try {
    if (editingId.value) {
      await api.put(`/providers/${editingId.value}`, form)
    } else {
      await api.post('/providers', form)
    }
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    fetchProviders()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  await api.delete(`/providers/${id}`)
  ElMessage.success(t('common.success'))
  fetchProviders()
}

async function handleBatchDelete() {
  if (selectedIds.value.length === 0) return
  await ElMessageBox.confirm(t('common.confirm') + ` (${selectedIds.value.length} items)`, t('common.batchDelete'), { type: 'warning' })
  try {
    await Promise.all(selectedIds.value.map(id => api.delete(`/providers/${id}`)))
    ElMessage.success(t('common.success'))
    selectedIds.value = []
    fetchProviders()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function toggleEnabled(row: any) {
  await api.put(`/providers/${row.id}`, { enabled: row.enabled })
}

function goDetail(id: number) {
  router.push(`/providers/${id}`)
}

function handleSortChange({ prop, order }: any) {
  if (prop && order) {
    setSortConfig('providers', { prop, order })
  }
}
</script>

<style scoped>
.providers {
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
</style>
