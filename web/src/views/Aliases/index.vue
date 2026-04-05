<template>
  <div class="aliases-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.aliases') }}</span>
          <div class="header-actions">
            <el-button type="danger" @click="handleBatchDelete" :disabled="selectedIds.length === 0">{{ t('common.batchDelete') }} ({{ selectedIds.length }})</el-button>
            <el-button type="primary" @click="showAliasDialog()">{{ t('modelAlias.create') }}</el-button>
          </div>
        </div>
      </template>

      <el-table :data="aliases" stripe v-loading="loading" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" :label="t('modelAlias.name')" width="220" />
        <el-table-column :label="t('modelAlias.mappingCount')" width="120">
          <template #default="{ row }">
            {{ row.mapping_count }}
          </template>
        </el-table-column>
        <el-table-column :label="t('modelAlias.capabilities')">
          <template #default="{ row }">
            <div v-if="row.mapping_count > 0" class="capability-tags">
              <el-tag v-if="row.supports_stream" type="primary" size="small" style="margin-right: 4px">Stream</el-tag>
              <el-tag v-if="row.supports_tools" type="warning" size="small" style="margin-right: 4px">Tools</el-tag>
              <el-tag v-if="row.supports_vision" type="success" size="small">Vision</el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('modelAlias.tokenSummary')">
          <template #default="{ row }">
            <el-tooltip v-if="row.min_context_window > 0 || row.min_max_output > 0" 
              :content="`${row.min_context_window.toLocaleString()} / ${row.min_max_output.toLocaleString()}`" 
              placement="top">
              <span>{{ formatContextDisplay(row.min_context_window, row.min_max_output) }}</span>
            </el-tooltip>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="120">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleAliasEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="showAliasDialog(row)">{{ t('common.edit') }}</el-button>
            <el-button link type="default" @click="goDetail(row.id)">{{ t('common.detail') }}</el-button>
            <el-button link type="danger" @click="handleDeleteAlias(row.id)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && aliases.length === 0" :description="t('common.noData')" />
    </el-card>

    <el-dialog v-model="aliasDialogVisible" :title="editingAlias ? t('modelAlias.editAlias') : t('modelAlias.createAlias')" width="400px">
      <el-form :model="aliasForm" :rules="aliasRules" ref="aliasFormRef" label-width="auto">
        <el-form-item :label="t('modelAlias.name')" prop="name">
          <el-input v-model="aliasForm.name" placeholder="e.g., gpt-4" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-switch v-model="aliasForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="aliasDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleAliasSubmit" :loading="submitting">{{ t('common.save') }}</el-button>
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
import { formatContextDisplay } from '@/utils/format'

const { t } = useI18n()
const router = useRouter()

interface Alias {
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

const aliases = ref<Alias[]>([])
const selectedIds = ref<number[]>([])
const loading = ref(false)
const aliasDialogVisible = ref(false)
const submitting = ref(false)
const editingAlias = ref<Alias | null>(null)
const aliasFormRef = ref()

const aliasForm = reactive({
  name: '',
  enabled: true
})

const aliasRules = computed(() => ({
  name: [{ required: true, message: t('common.required'), trigger: 'blur' }]
}))

onMounted(() => {
  fetchAliases()
})

async function fetchAliases() {
  loading.value = true
  try {
    const res = await api.get('/aliases')
    aliases.value = (res.data.aliases || []).map((a: any) => ({
      id: a.id,
      name: a.alias,
      enabled: a.enabled,
      mapping_count: a.mapping_count,
      min_context_window: a.min_context_window || 0,
      min_max_output: a.min_max_output || 0,
      supports_vision: a.supports_vision || false,
      supports_tools: a.supports_tools || false,
      supports_stream: a.supports_stream || false
    }))
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(selection: Alias[]) {
  selectedIds.value = selection.map(item => item.id)
}

async function showAliasDialog(alias?: Alias) {
  editingAlias.value = alias || null
  Object.assign(aliasForm, {
    name: alias?.name || '',
    enabled: alias?.enabled ?? true
  })
  aliasDialogVisible.value = true
}

async function handleAliasSubmit() {
  const valid = await aliasFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingAlias.value) {
      await api.put(`/aliases/${editingAlias.value.id}`, { name: aliasForm.name, enabled: aliasForm.enabled })
    } else {
      await api.post('/aliases', { name: aliasForm.name, enabled: aliasForm.enabled })
    }
    ElMessage.success(t('common.success'))
    aliasDialogVisible.value = false
    fetchAliases()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    submitting.value = false
  }
}

async function handleDeleteAlias(id: number) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  await api.delete(`/aliases/${id}`)
  ElMessage.success(t('common.success'))
  fetchAliases()
}

async function handleBatchDelete() {
  if (selectedIds.value.length === 0) return
  await ElMessageBox.confirm(t('common.confirm') + ` (${selectedIds.value.length} items)`, t('common.batchDelete'), { type: 'warning' })
  try {
    await Promise.all(selectedIds.value.map(id => api.delete(`/aliases/${id}`)))
    ElMessage.success(t('common.success'))
    selectedIds.value = []
    fetchAliases()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function toggleAliasEnabled(alias: Alias) {
  await api.put(`/aliases/${alias.id}`, { enabled: alias.enabled })
}

function goDetail(id: number) {
  router.push(`/aliases/${id}`)
}
</script>

<style scoped>
.aliases-page {
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
</style>