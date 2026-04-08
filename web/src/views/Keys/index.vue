<template>
  <div class="keys-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.keys') }}</span>
          <div class="header-actions">
            <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">{{ t('common.batchDelete') }} ({{ selectedIds.length }})</el-button>
            <el-button type="primary" @click="showDialog()">{{ t('key.createKey') }}</el-button>
          </div>
        </div>
      </template>
      <el-table :data="keys" stripe v-loading="loading" @selection-change="handleSelectionChange" :default-sort="defaultSort" @sort-change="handleSortChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" :label="t('key.name')" width="180" sortable />
        <el-table-column prop="key" :label="t('key.key')" width="240" />
        <el-table-column :label="t('key.model')" prop="models" sortable :sort-method="(a: any, b: any) => sortByArrayLength(a, b, 'models')">
          <template #default="{ row }">
            <span v-if="!row.models || row.models.length === 0" style="color: #999">{{ t('key.allModels') }}</span>
            <span v-else>{{ t('key.allowedCount', { count: row.models.length }) }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('mcp.tools')" prop="mcp_tools_count" sortable>
          <template #default="{ row }">
            <span v-if="row.mcp_tools_count === 0" style="color: #999">{{ t('key.allModels') }}</span>
            <span v-else>{{ t('key.allowedCount', { count: row.mcp_tools_count }) }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('mcp.resources')" prop="mcp_resources_count" sortable>
          <template #default="{ row }">
            <span v-if="row.mcp_resources_count === 0" style="color: #999">{{ t('key.allModels') }}</span>
            <span v-else>{{ t('key.allowedCount', { count: row.mcp_resources_count }) }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('mcp.prompts')" prop="mcp_prompts_count" sortable>
          <template #default="{ row }">
            <span v-if="row.mcp_prompts_count === 0" style="color: #999">{{ t('key.allModels') }}</span>
            <span v-else>{{ t('key.allowedCount', { count: row.mcp_prompts_count }) }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="150" prop="enabled" sortable>
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="180">
           <template #default="{ row }">
             <el-button link type="primary" @click="showDialog(row)">{{ t('common.edit') }}</el-button>
             <el-button link type="default" @click="goDetail(row.id)">{{ t('common.detail') }}</el-button>
             <el-button link type="danger" @click="handleDelete(row.id)">{{ t('common.delete') }}</el-button>
           </template>
         </el-table-column>
      </el-table>
    </el-card>

<el-dialog v-model="dialogVisible" :title="editingId ? t('common.edit') : t('key.createKey')" width="500px">
       <el-form :model="form" ref="formRef" label-width="auto">
         <el-form-item :label="t('key.name')" required>
           <el-input v-model="form.name" />
         </el-form-item>
       </el-form>
       <template #footer>
         <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
         <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ t('common.save') }}</el-button>
       </template>
     </el-dialog>

    <el-dialog v-model="keyDialogVisible" title="API Key">
      <p>{{ t('key.key') }}:</p>
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
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api'
import { getSortConfig, setSortConfig, sortByArrayLength } from '@/utils/tableSort'

const { t } = useI18n()
const router = useRouter()

const keys = ref<any[]>([])
const selectedIds = ref<number[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const keyDialogVisible = ref(false)
const newKey = ref('')
const editingId = ref<number | null>(null)
const formRef = ref()
const submitting = ref(false)
const defaultSort = getSortConfig('keys', 'name')

const form = reactive({
  name: ''
})

onMounted(() => {
  fetchKeys()
})

async function fetchKeys() {
  loading.value = true
  try {
    const res = await api.get('/keys')
    keys.value = res.data.keys || []
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(selection: any[]) {
  selectedIds.value = selection.map(item => item.id)
}

async function showDialog(key?: any) {
  editingId.value = key?.id || null
  
  if (key) {
    form.name = key.name || ''
  } else {
    form.name = ''
  }
  
  dialogVisible.value = true
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (editingId.value) {
      await api.put(`/keys/${editingId.value}`, { name: form.name })
      ElMessage.success(t('common.success'))
      dialogVisible.value = false
    } else {
      const res = await api.post('/keys', { name: form.name })
      newKey.value = res.data.raw_key
      dialogVisible.value = false
      keyDialogVisible.value = true
    }
    fetchKeys()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    submitting.value = false
  }
}

async function toggleEnabled(row: any) {
  await api.put(`/keys/${row.id}`, { enabled: row.enabled })
}

function copyKey() {
  navigator.clipboard.writeText(newKey.value)
  ElMessage.success('Copied!')
}

function goDetail(id: number) {
  router.push(`/keys/${id}`)
}

function handleSortChange({ prop, order }: any) {
  if (prop && order) {
    setSortConfig('keys', { prop, order })
  }
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  await api.delete(`/keys/${id}`)
  ElMessage.success(t('common.success'))
  fetchKeys()
}

async function handleBatchDelete() {
  if (selectedIds.value.length === 0) return
  await ElMessageBox.confirm(t('common.confirm') + ` (${selectedIds.value.length} items)`, t('common.batchDelete'), { type: 'warning' })
  try {
    await Promise.all(selectedIds.value.map(id => api.delete(`/keys/${id}`)))
    ElMessage.success(t('common.success'))
    selectedIds.value = []
    fetchKeys()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}
</script>

<style scoped>
.keys-page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; gap: 10px; }
</style>
