<template>
  <div class="mcp-services">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('mcp.services') }}</span>
          <div>
            <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">{{ t('common.batchDelete') }} ({{ selectedIds.length }})</el-button>
            <el-button type="primary" @click="showDialog()">{{ t('mcp.addService') }}</el-button>
          </div>
        </div>
      </template>
      <el-table :data="services" stripe v-loading="loading" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" :label="t('mcp.name')" width="200" />
        <el-table-column :label="t('mcp.type')">
          <template #default="{ row }">
            <el-tag :type="row.type === 'remote' ? 'success' : 'info'">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('mcp.tools')">
          <template #default="{ row }">
            {{ row.tool_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('mcp.resources')">
          <template #default="{ row }">
            {{ row.resource_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('mcp.prompts')">
          <template #default="{ row }">
            {{ row.prompt_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column prop="last_sync_at" :label="t('mcp.lastSync')" width="180">
          <template #default="{ row }">
            {{ row.last_sync_at ? new Date(row.last_sync_at).toLocaleString() : '-' }}
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

    <el-dialog v-model="dialogVisible" :title="editingId ? t('mcp.editService') : t('mcp.addService')" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" v-loading="dialogLoading">
        <el-form-item :label="t('mcp.name')" prop="name">
          <el-input v-model="form.name" :placeholder="t('mcp.namePlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('mcp.type')" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio label="remote">{{ t('mcp.remote') }}</el-radio>
            <el-radio label="local">{{ t('mcp.local') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item :label="form.type === 'remote' ? t('mcp.url') : t('mcp.command')" prop="target">
          <el-input v-model="form.target" :placeholder="form.type === 'remote' ? t('mcp.urlPlaceholder') : t('mcp.commandPlaceholder')" />
        </el-form-item>
        <el-form-item :label="form.type === 'remote' ? t('mcp.headers') : t('mcp.envVars')">
          <el-input v-model="form.params" type="textarea" :rows="5" />
          <div class="form-tip">{{ form.type === 'remote' ? t('mcp.paramsRemotePlaceholder') : t('mcp.paramsLocalPlaceholder') }}</div>
        </el-form-item>
        
        <el-form-item :label="t('common.status')">
          <el-switch v-model="form.enabled" />
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

const services = ref<any[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const dialogLoading = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref()
const selectedIds = ref<number[]>([])

const form = reactive({
  name: '',
  type: 'local',
  target: '',
  params: '',
  enabled: true
})

const rules = computed(() => ({
  name: [
    { required: true, message: 'Required', trigger: 'blur' },
    { pattern: /^[0-9a-zA-Z_-]{2,200}$/, message: t('mcp.nameInvalid'), trigger: 'blur' }
  ],
  type: [{ required: true, message: 'Required', trigger: 'change' }],
  target: [{ required: true, message: 'Required', trigger: 'blur' }]
}))

onMounted(() => {
  fetchServices()
})

async function fetchServices() {
  loading.value = true
  try {
    const res = await api.get('/mcps')
    services.value = res.data.mcps || []
  } finally {
    loading.value = false
  }
}

function goDetail(id: number) {
  router.push(`/mcps/${id}`)
}

async function showDialog(id?: number) {
  editingId.value = id || null
  Object.assign(form, {
    name: '',
    type: 'local',
    target: '',
    params: '',
    enabled: true
  })
  dialogVisible.value = true
  
  if (id) {
    dialogLoading.value = true
    try {
      const res = await api.get(`/mcps/${id}`)
      const mcp = res.data.mcp
      if (mcp) {
        Object.assign(form, {
          name: mcp.name || '',
          type: mcp.type || 'local',
          target: mcp.target || '',
          params: mcp.params || '',
          enabled: mcp.enabled ?? true
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
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    
    submitting.value = true
    try {
      if (editingId.value) {
        await api.put(`/mcps/${editingId.value}`, form)
        ElMessage.success(t('common.success'))
      } else {
        await api.post('/mcps', form)
        ElMessage.success(t('common.success'))
      }
      dialogVisible.value = false
      fetchServices()
    } catch (e: any) {
      ElMessage.error(e.response?.data?.error || t('common.error'))
    } finally {
      submitting.value = false
    }
  })
}

async function toggleEnabled(row: any) {
  try {
    await api.put(`/mcps/${row.id}`, { enabled: row.enabled })
    ElMessage.success(t('common.success'))
  } catch (e) {
    row.enabled = !row.enabled
    ElMessage.error(t('common.error'))
  }
}

async function testConnection(id: number) {
  try {
    const res = await api.post(`/mcps/${id}/test`)
    if (res.data.success) {
      ElMessage.success(t('mcp.connectionSuccess'))
    } else {
      ElMessage.error(res.data.message || t('mcp.connectionFailed'))
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('mcp.connectionFailed'))
  }
}

async function handleDelete(id: number) {
  try {
    await ElMessageBox.confirm(t('common.confirmDelete'), t('common.warning'), {
      type: 'warning'
    })
    await api.delete(`/mcps/${id}`)
    ElMessage.success(t('common.success'))
    fetchServices()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.error || t('common.error'))
    }
  }
}

function handleSelectionChange(selection: any[]) {
  selectedIds.value = selection.map(item => item.id)
}

async function handleBatchDelete() {
  if (selectedIds.value.length === 0) return
  
  try {
    await ElMessageBox.confirm(
      t('mcp.confirmBatchDelete', { count: selectedIds.value.length }),
      t('common.warning'),
      { type: 'warning' }
    )
    
    await Promise.all(selectedIds.value.map(id => api.delete(`/mcps/${id}`)))
    ElMessage.success(t('common.success'))
    selectedIds.value = []
    fetchServices()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.error || t('common.error'))
    }
  }
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}
</style>
