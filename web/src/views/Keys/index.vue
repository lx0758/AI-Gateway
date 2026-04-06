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
      <el-table :data="keys" stripe v-loading="loading" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" :label="t('key.name')" width="180" />
        <el-table-column prop="key" :label="t('key.key')" width="240" />
        <el-table-column :label="t('key.allowedModels')">
          <template #default="{ row }">
            <template v-if="row.models && row.models.length > 0">
              <el-tag v-for="m in row.models.slice(0, 3)" :key="m.id" size="small" style="margin-right: 4px">
                {{ m.model_name }}
              </el-tag>
              <el-tag v-if="row.models.length > 3" size="small" type="info">+{{ row.models.length - 3 }}</el-tag>
            </template>
            <span v-else style="color: #999">{{ t('key.allModels') }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="150">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDialog(row)">{{ t('common.edit') }}</el-button>
            <el-button link type="warning" @click="handleReset(row.id)">{{ t('key.reset') }}</el-button>
            <el-button link type="danger" @click="handleDelete(row.id)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? t('common.edit') : t('key.createKey')" width="700px">
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="t('key.allowedModels')" name="models">
          <el-form :model="form" ref="formRef" label-width="auto">
            <el-form-item :label="t('key.name')" required>
              <el-input v-model="form.name" />
            </el-form-item>
            <el-form-item :label="t('key.allowedModels')">
              <el-select v-model="form.models" multiple style="width: 100%" :placeholder="t('key.allModels')" filterable>
                <el-option v-for="m in availableModels" :key="m.id" :label="m.name" :value="m.id" />
              </el-select>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane :label="t('mcp.tools')" name="mcpTools">
          <div v-loading="mcpToolsLoading">
            <el-transfer
              v-model="selectedMCPTools"
              :data="availableMCPTools"
              :titles="[t('common.available'), t('common.selected')]"
              :props="{ key: 'id', label: 'displayName' }"
              filterable
              :filter-placeholder="t('common.search')"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('mcp.resources')" name="mcpResources">
          <div v-loading="mcpResourcesLoading">
            <el-transfer
              v-model="selectedMCPResources"
              :data="availableMCPResources"
              :titles="[t('common.available'), t('common.selected')]"
              :props="{ key: 'id', label: 'displayName' }"
              filterable
              :filter-placeholder="t('common.search')"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('mcp.prompts')" name="mcpPrompts">
          <div v-loading="mcpPromptsLoading">
            <el-transfer
              v-model="selectedMCPPrompts"
              :data="availableMCPPrompts"
              :titles="[t('common.available'), t('common.selected')]"
              :props="{ key: 'id', label: 'displayName' }"
              filterable
              :filter-placeholder="t('common.search')"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
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
const submitting = ref(false)
const activeTab = ref('models')
let modelsLoaded = false

const availableMCPTools = ref<any[]>([])
const availableMCPResources = ref<any[]>([])
const availableMCPPrompts = ref<any[]>([])
const selectedMCPTools = ref<number[]>([])
const selectedMCPResources = ref<number[]>([])
const selectedMCPPrompts = ref<number[]>([])
const mcpToolsLoading = ref(false)
const mcpResourcesLoading = ref(false)
const mcpPromptsLoading = ref(false)

const form = reactive({
  name: '',
  models: [] as number[]
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

async function fetchAvailableModels() {
  try {
    const res = await api.get('/models')
    availableModels.value = (res.data.models || []).map((m: any) => ({ id: m.id, name: m.model }))
  } catch (e) {
    console.error(e)
  }
}

function handleSelectionChange(selection: any[]) {
  selectedIds.value = selection.map(item => item.id)
}

async function showDialog(key?: any) {
  if (!modelsLoaded) {
    await fetchAvailableModels()
    modelsLoaded = true
  }
  editingId.value = key?.id || null
  activeTab.value = 'models'
  
  if (key) {
    Object.assign(form, {
      name: key.name || '',
      models: key.models?.map((m: any) => m.model_id) || []
    })
    
    // Load MCP permissions
    loadMCPPermissions(key.id)
  } else {
    Object.assign(form, { name: '', models: [] })
    selectedMCPTools.value = []
    selectedMCPResources.value = []
    selectedMCPPrompts.value = []
  }
  
  // Load available MCP resources
  fetchAvailableMCPTools()
  fetchAvailableMCPResources()
  fetchAvailableMCPPrompts()
  
  dialogVisible.value = true
}

async function loadMCPPermissions(keyId: number) {
  try {
    const [toolsRes, resourcesRes, promptsRes] = await Promise.all([
      api.get(`/keys/${keyId}/mcp-tools`),
      api.get(`/keys/${keyId}/mcp-resources`),
      api.get(`/keys/${keyId}/mcp-prompts`)
    ])
    
    selectedMCPTools.value = (toolsRes.data.tools || []).map((t: any) => t.tool_id)
    selectedMCPResources.value = (resourcesRes.data.resources || []).map((r: any) => r.resource_id)
    selectedMCPPrompts.value = (promptsRes.data.prompts || []).map((p: any) => p.prompt_id)
  } catch (e) {
    console.error('Failed to load MCP permissions', e)
  }
}

async function fetchAvailableMCPTools() {
  mcpToolsLoading.value = true
  try {
    const res = await api.get('/mcps')
    const mcps = res.data.mcps || []
    
    const tools: any[] = []
    for (const mcp of mcps) {
      if (!mcp.enabled) continue
      const toolsRes = await api.get(`/mcps/${mcp.id}/tools`)
      const mcpTools = toolsRes.data.tools || []
      tools.push(...mcpTools.filter((t: any) => t.enabled).map((t: any) => ({
        id: t.id,
        displayName: `${mcp.name}.${t.name}`
      })))
    }
    availableMCPTools.value = tools
  } finally {
    mcpToolsLoading.value = false
  }
}

async function fetchAvailableMCPResources() {
  mcpResourcesLoading.value = true
  try {
    const res = await api.get('/mcps')
    const mcps = res.data.mcps || []
    
    const resources: any[] = []
    for (const mcp of mcps) {
      if (!mcp.enabled) continue
      const resourcesRes = await api.get(`/mcps/${mcp.id}/resources`)
      const mcpResources = resourcesRes.data.resources || []
      resources.push(...mcpResources.filter((r: any) => r.enabled).map((r: any) => ({
        id: r.id,
        displayName: `${mcp.name}.${r.name || r.uri}`
      })))
    }
    availableMCPResources.value = resources
  } finally {
    mcpResourcesLoading.value = false
  }
}

async function fetchAvailableMCPPrompts() {
  mcpPromptsLoading.value = true
  try {
    const res = await api.get('/mcps')
    const mcps = res.data.mcps || []
    
    const prompts: any[] = []
    for (const mcp of mcps) {
      if (!mcp.enabled) continue
      const promptsRes = await api.get(`/mcps/${mcp.id}/prompts`)
      const mcpPrompts = promptsRes.data.prompts || []
      prompts.push(...mcpPrompts.filter((p: any) => p.enabled).map((p: any) => ({
        id: p.id,
        displayName: `${mcp.name}.${p.name}`
      })))
    }
    availableMCPPrompts.value = prompts
  } finally {
    mcpPromptsLoading.value = false
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (editingId.value) {
      await Promise.all([
        api.put(`/keys/${editingId.value}`, form),
        api.put(`/keys/${editingId.value}/mcp-tools`, { tool_ids: selectedMCPTools.value }),
        api.put(`/keys/${editingId.value}/mcp-resources`, { resource_ids: selectedMCPResources.value }),
        api.put(`/keys/${editingId.value}/mcp-prompts`, { prompt_ids: selectedMCPPrompts.value })
      ])
      ElMessage.success(t('common.success'))
      dialogVisible.value = false
    } else {
      const res = await api.post('/keys', form)
      const keyId = res.data.key.id
      
      // Save MCP permissions for new key
      if (selectedMCPTools.value.length > 0 || selectedMCPResources.value.length > 0 || selectedMCPPrompts.value.length > 0) {
        await Promise.all([
          api.put(`/keys/${keyId}/mcp-tools`, { tool_ids: selectedMCPTools.value }),
          api.put(`/keys/${keyId}/mcp-resources`, { resource_ids: selectedMCPResources.value }),
          api.put(`/keys/${keyId}/mcp-prompts`, { prompt_ids: selectedMCPPrompts.value })
        ])
      }
      
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

async function handleDelete(id: number) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  await api.delete(`/keys/${id}`)
  ElMessage.success(t('common.success'))
  fetchKeys()
}

async function handleReset(id: number) {
  await ElMessageBox.confirm(
    t('key.resetConfirmMessage'),
    t('key.resetConfirmTitle'),
    { type: 'warning' }
  )
  try {
    const res = await api.post(`/keys/${id}/reset`)
    newKey.value = res.data.raw_key
    keyDialogVisible.value = true
    fetchKeys()
    ElMessage.success(t('key.resetSuccess'))
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
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
