<template>
  <div class="providers">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.providers') }}</span>
          <el-button type="primary" @click="showDialog()">
            {{ t('provider.addProvider') }}
          </el-button>
        </div>
      </template>
      <el-table :data="providers" stripe v-loading="loading">
        <el-table-column prop="name" :label="t('provider.name')" />
        <el-table-column prop="api_type" :label="t('provider.apiType')">
          <template #default="{ row }">
            {{ formatApiType(row.api_type) }}
          </template>
        </el-table-column>
        <el-table-column prop="base_url" :label="t('provider.baseUrl')" />
        <el-table-column :label="t('provider.models')">
          <template #default="{ row }">
            {{ row.models?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="250">
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
        <el-form-item :label="t('provider.apiType')" prop="api_type">
          <el-select v-model="form.api_type" style="width: 100%">
            <el-option label="@ai-sdk/openai-compatible" value="openai" />
            <el-option label="@ai-sdk/anthropic" value="anthropic" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('provider.baseUrl')" prop="base_url">
          <el-input v-model="form.base_url" />
        </el-form-item>
        <el-form-item :label="t('provider.apiKey')" prop="api_key">
          <el-input v-model="form.api_key" type="password" show-password :placeholder="editingId ? t('provider.apiKeyPlaceholder') : ''" />
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

const { t } = useI18n()
const router = useRouter()

const providers = ref<any[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogLoading = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref()

const form = reactive({
  name: '',
  api_type: 'openai',
  base_url: '',
  api_key: ''
})

const rules = computed(() => ({
  name: [{ required: true, message: 'Required', trigger: 'blur' }],
  api_type: [{ required: true, message: 'Required', trigger: 'change' }],
  base_url: [{ required: true, message: 'Required', trigger: 'blur' }],
  api_key: editingId.value ? [] : [{ required: true, message: 'Required', trigger: 'blur' }]
}))

const apiTypeLabels: Record<string, string> = {
  openai: '@ai-sdk/openai-compatible',
  anthropic: '@ai-sdk/anthropic'
}

function formatApiType(type: string) {
  return apiTypeLabels[type] || type
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

async function showDialog(id?: number) {
  editingId.value = id || null
  Object.assign(form, { name: '', api_type: 'openai', base_url: '', api_key: '' })
  dialogVisible.value = true
  
  if (id) {
    dialogLoading.value = true
    try {
      const res = await api.get(`/providers/${id}`)
      const provider = res.data.provider
      if (provider) {
        Object.assign(form, {
          name: provider.name || '',
          api_type: provider.api_type || 'openai',
          base_url: provider.base_url || '',
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

function goDetail(id: number) {
  router.push(`/providers/${id}`)
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
</style>
