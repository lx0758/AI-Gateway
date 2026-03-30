<template>
  <div class="api-keys-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.apiKeys') }}</span>
          <el-button type="primary" @click="showDialog()">{{ t('apiKey.createKey') }}</el-button>
        </div>
      </template>
      <el-table :data="keys" stripe v-loading="loading">
        <el-table-column prop="name" :label="t('apiKey.name')" />
        <el-table-column prop="key" :label="t('apiKey.key')" />
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
        <el-table-column :label="t('apiKey.usedQuota')">
          <template #default="{ row }">
            {{ row.used_quota }} / {{ row.quota || '∞' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'">
              {{ row.enabled ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="100">
          <template #default="{ row }">
            <el-button link type="danger" @click="handleDelete(row.id)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="t('apiKey.createKey')">
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

const { t } = useI18n()

const keys = ref<any[]>([])
const availableModels = ref<any[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const keyDialogVisible = ref(false)
const newKey = ref('')
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

function showDialog() {
  Object.assign(form, { name: '', quota: 0, rate_limit: 0, allowed_models: [] })
  dialogVisible.value = true
}

async function handleSubmit() {
  try {
    const res = await api.post('/api-keys', form)
    newKey.value = res.data.raw_key
    dialogVisible.value = false
    keyDialogVisible.value = true
    fetchKeys()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
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
</script>

<style scoped>
.api-keys-page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
</style>
