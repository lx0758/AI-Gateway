<template>
  <div class="key-detail">
    <el-page-header @back="goBack" :title="t('menu.keys')">
      <template #content>
        <span class="text-large font-600 mr-3">{{ key?.name || '-' }}</span>
      </template>
    </el-page-header>

    <el-card class="info-card" v-loading="loading">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="t('key.name')">{{ key?.name }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.createdAt')">
          {{ key?.created_at ? new Date(key.created_at).toLocaleString() : '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">
          <el-tag :type="key?.enabled ? 'success' : 'info'" size="small">
            {{ key?.enabled ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('key.expiresAt')">
          {{ key?.expires_at ? new Date(key.expires_at).toLocaleString() : '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('key.key')">
          <code>{{ key?.key }}</code>
        </el-descriptions-item>
      </el-descriptions>
      <div class="actions">
        <el-button @click="handleReset" :loading="resetting">{{ t('key.reset') }}</el-button>
      </div>
    </el-card>

    <el-card class="tabs-card">
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="t('key.model')" name="models">
          <div class="tab-actions">
            <el-button @click="clearModels" :loading="clearingModels" :disabled="!key?.enabled">{{ t('key.allowAll') }}</el-button>
          </div>
          <el-table :data="models" stripe v-loading="modelsLoading" :default-sort="modelsDefaultSort" @sort-change="(e: any) => handleSortChange('key-models', e)">
            <el-table-column prop="name" :label="t('models.name')" width="200" sortable />
            <el-table-column :label="t('models.mappingCount')" width="120" prop="mapping_count" sortable>
              <template #default="{ row }">
                {{ row.mapping_count || 0 }}
              </template>
            </el-table-column>
            <el-table-column :label="t('models.capabilities')">
              <template #default="{ row }">
                <div class="capability-tags">
                  <el-tag v-if="row.supports_stream" type="primary" size="small" style="margin-right: 4px">Stream</el-tag>
                  <el-tag v-if="row.supports_tools" type="warning" size="small" style="margin-right: 4px">Tools</el-tag>
                  <el-tag v-if="row.supports_vision" type="success" size="small">Vision</el-tag>
                  <span v-if="!row.supports_stream && !row.supports_tools && !row.supports_vision">-</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column :label="t('models.contextWindow')" width="150" prop="min_context_window" sortable>
              <template #default="{ row }">
                <el-tooltip v-if="row.min_context_window > 0 || row.min_max_output > 0"
                  :content="`${row.min_context_window?.toLocaleString() || 0} / ${row.min_max_output?.toLocaleString() || 0}`"
                  placement="top">
                  <span>{{ formatContextDisplay(row.min_context_window, row.min_max_output) }}</span>
                </el-tooltip>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column :label="t('common.status')" width="180" prop="selected" sortable>
              <template #default="{ row }">
                <el-tooltip v-if="!key?.enabled" :content="t('key.keyDisabled')" placement="top">
                  <el-radio-group v-model="row.selected" disabled>
                    <el-radio :label="false">{{ t('key.default') }}</el-radio>
                    <el-radio :label="true">{{ t('key.allowOnly') }}</el-radio>
                  </el-radio-group>
                </el-tooltip>
                <el-radio-group v-else v-model="row.selected" @change="toggleModel(row)">
                  <el-radio :label="false">{{ t('key.default') }}</el-radio>
                  <el-radio :label="true">{{ t('key.allowOnly') }}</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('mcp.tools')" name="tools">
          <div class="tab-actions">
            <el-button @click="clearTools" :loading="clearingTools" :disabled="!key?.enabled">{{ t('key.allowAll') }}</el-button>
          </div>
          <el-table :data="tools" stripe v-loading="toolsLoading" :default-sort="toolsDefaultSort" @sort-change="(e: any) => handleSortChange('key-tools', e)">
            <el-table-column :label="t('mcp.toolName')" width="300" prop="name" sortable>
              <template #default="{ row }">
                <div>{{ row.mcp_name }}.{{ row.name }}</div>
              </template>
            </el-table-column>
            <el-table-column :label="t('mcp.description')" prop="description" sortable>
              <template #default="{ row }">
                <div class="description-cell">
                  <div class="description-text" :class="{ expanded: row._expanded }">
                    {{ row.description || '-' }}
                  </div>
                  <div class="description-actions">
                    <el-button v-if="row.description && isLongText(row.description)" link type="primary" size="small" @click="row._expanded = !row._expanded">
                      {{ row._expanded ? t('common.collapse') : t('common.expand') }}
                    </el-button>
                    <el-button v-if="row.description" link type="primary" size="small" @click="copyText(row.description)">
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column :label="t('common.status')" width="180" prop="selected" sortable>
              <template #default="{ row }">
                <el-tooltip v-if="!key?.enabled" :content="t('key.keyDisabled')" placement="top">
                  <el-radio-group v-model="row.selected" disabled>
                    <el-radio :label="false">{{ t('key.default') }}</el-radio>
                    <el-radio :label="true">{{ t('key.allowOnly') }}</el-radio>
                  </el-radio-group>
                </el-tooltip>
                <el-radio-group v-else v-model="row.selected" @change="toggleTool(row)">
                  <el-radio :label="false">{{ t('key.default') }}</el-radio>
                  <el-radio :label="true">{{ t('key.allowOnly') }}</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('mcp.resources')" name="resources">
          <div class="tab-actions">
            <el-button @click="clearResources" :loading="clearingResources" :disabled="!key?.enabled">{{ t('key.allowAll') }}</el-button>
          </div>
          <el-table :data="resources" stripe v-loading="resourcesLoading" :default-sort="resourcesDefaultSort" @sort-change="(e: any) => handleSortChange('key-resources', e)">
            <el-table-column :label="t('mcp.resourceName')" width="300" prop="name" sortable>
              <template #default="{ row }">
                <div>{{ row.mcp_name }}.{{ row.name }}</div>
              </template>
            </el-table-column>
            <el-table-column :label="t('mcp.description')" prop="description" sortable>
              <template #default="{ row }">
                <div class="description-cell">
                  <div class="description-text" :class="{ expanded: row._expanded }">
                    {{ row.description || '-' }}
                  </div>
                  <div class="description-actions">
                    <el-button v-if="row.description && isLongText(row.description)" link type="primary" size="small" @click="row._expanded = !row._expanded">
                      {{ row._expanded ? t('common.collapse') : t('common.expand') }}
                    </el-button>
                    <el-button v-if="row.description" link type="primary" size="small" @click="copyText(row.description)">
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="uri" :label="t('mcp.resourceUri')" width="300" sortable />
            <el-table-column prop="mime_type" label="MIME Type" width="200" sortable />
            <el-table-column :label="t('common.status')" width="180" prop="selected" sortable>
              <template #default="{ row }">
                <el-tooltip v-if="!key?.enabled" :content="t('key.keyDisabled')" placement="top">
                  <el-radio-group v-model="row.selected" disabled>
                    <el-radio :label="false">{{ t('key.default') }}</el-radio>
                    <el-radio :label="true">{{ t('key.allowOnly') }}</el-radio>
                  </el-radio-group>
                </el-tooltip>
                <el-radio-group v-else v-model="row.selected" @change="toggleResource(row)">
                  <el-radio :label="false">{{ t('key.default') }}</el-radio>
                  <el-radio :label="true">{{ t('key.allowOnly') }}</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('mcp.prompts')" name="prompts">
          <div class="tab-actions">
            <el-button @click="clearPrompts" :loading="clearingPrompts" :disabled="!key?.enabled">{{ t('key.allowAll') }}</el-button>
          </div>
          <el-table :data="prompts" stripe v-loading="promptsLoading" :default-sort="promptsDefaultSort" @sort-change="(e: any) => handleSortChange('key-prompts', e)">
            <el-table-column :label="t('mcp.promptName')" width="300" prop="name" sortable>
              <template #default="{ row }">
                <div>{{ row.mcp_name }}.{{ row.name }}</div>
              </template>
            </el-table-column>
            <el-table-column :label="t('mcp.description')" prop="description" sortable>
              <template #default="{ row }">
                <div class="description-cell">
                  <div class="description-text" :class="{ expanded: row._expanded }">
                    {{ row.description || '-' }}
                  </div>
                  <div class="description-actions">
                    <el-button v-if="row.description && isLongText(row.description)" link type="primary" size="small" @click="row._expanded = !row._expanded">
                      {{ row._expanded ? t('common.collapse') : t('common.expand') }}
                    </el-button>
                    <el-button v-if="row.description" link type="primary" size="small" @click="copyText(row.description)">
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column :label="t('common.status')" width="180" prop="selected" sortable>
              <template #default="{ row }">
                <el-tooltip v-if="!key?.enabled" :content="t('key.keyDisabled')" placement="top">
                  <el-radio-group v-model="row.selected" disabled>
                    <el-radio :label="false">{{ t('key.default') }}</el-radio>
                    <el-radio :label="true">{{ t('key.allowOnly') }}</el-radio>
                  </el-radio-group>
                </el-tooltip>
                <el-radio-group v-else v-model="row.selected" @change="togglePrompt(row)">
                  <el-radio :label="false">{{ t('key.default') }}</el-radio>
                  <el-radio :label="true">{{ t('key.allowOnly') }}</el-radio>
                </el-radio-group>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>

  <el-dialog v-model="keyDialogVisible" :title="t('key.key')" width="500px">
    <p>{{ t('key.key') }}:</p>
    <el-input v-model="newKey" readonly>
      <template #append>
        <el-button @click="copyKey">{{ t('common.copied') }}</el-button>
      </template>
    </el-input>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import api from '@/api'
import { formatContextDisplay } from '@/utils/format'
import { getSortConfig, setSortConfig } from '@/utils/tableSort'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const keyId = Number(route.params.id)
const key = ref<any>(null)
const loading = ref(false)
const resetting = ref(false)
const activeTab = ref('models')
const keyDialogVisible = ref(false)
const newKey = ref('')

const models = ref<any[]>([])
const tools = ref<any[]>([])
const resources = ref<any[]>([])
const prompts = ref<any[]>([])

const modelsLoading = ref(false)
const toolsLoading = ref(false)
const resourcesLoading = ref(false)
const promptsLoading = ref(false)

const clearingModels = ref(false)
const clearingTools = ref(false)
const clearingResources = ref(false)
const clearingPrompts = ref(false)

const modelsDefaultSort = getSortConfig('key-models', 'name')
const toolsDefaultSort = getSortConfig('key-tools', 'name')
const resourcesDefaultSort = getSortConfig('key-resources', 'name')
const promptsDefaultSort = getSortConfig('key-prompts', 'name')

onMounted(() => {
  fetchKey()
})

watch(activeTab, (newTab) => {
  if (newTab === 'models' && models.value.length === 0) fetchModels()
  if (newTab === 'tools' && tools.value.length === 0) fetchTools()
  if (newTab === 'resources' && resources.value.length === 0) fetchResources()
  if (newTab === 'prompts' && prompts.value.length === 0) fetchPrompts()
})

async function fetchKey() {
  loading.value = true
  try {
    const res = await api.get(`/keys/${keyId}`)
    key.value = res.data.key
    fetchModels()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
    goBack()
  } finally {
    loading.value = false
  }
}

async function fetchModels() {
  modelsLoading.value = true
  try {
    const res = await api.get(`/keys/${keyId}/models`)
    models.value = res.data.models || []
  } finally {
    modelsLoading.value = false
  }
}

async function fetchTools() {
  toolsLoading.value = true
  try {
    const res = await api.get(`/keys/${keyId}/mcp-tools`)
    tools.value = res.data.tools || []
  } finally {
    toolsLoading.value = false
  }
}

async function fetchResources() {
  resourcesLoading.value = true
  try {
    const res = await api.get(`/keys/${keyId}/mcp-resources`)
    resources.value = res.data.resources || []
  } finally {
    resourcesLoading.value = false
  }
}

async function fetchPrompts() {
  promptsLoading.value = true
  try {
    const res = await api.get(`/keys/${keyId}/mcp-prompts`)
    prompts.value = res.data.prompts || []
  } finally {
    promptsLoading.value = false
  }
}

async function toggleModel(row: any) {
  const previousValue = !row.selected
  try {
    if (row.selected) {
      await api.post(`/keys/${keyId}/models/${row.id}`)
    } else {
      await api.delete(`/keys/${keyId}/models/${row.id}`)
    }
  } catch (e: any) {
    row.selected = previousValue
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function toggleTool(row: any) {
  const previousValue = !row.selected
  try {
    if (row.selected) {
      await api.post(`/keys/${keyId}/mcp-tools/${row.id}`)
    } else {
      await api.delete(`/keys/${keyId}/mcp-tools/${row.id}`)
    }
  } catch (e: any) {
    row.selected = previousValue
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function toggleResource(row: any) {
  const previousValue = !row.selected
  try {
    if (row.selected) {
      await api.post(`/keys/${keyId}/mcp-resources/${row.id}`)
    } else {
      await api.delete(`/keys/${keyId}/mcp-resources/${row.id}`)
    }
  } catch (e: any) {
    row.selected = previousValue
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function togglePrompt(row: any) {
  const previousValue = !row.selected
  try {
    if (row.selected) {
      await api.post(`/keys/${keyId}/mcp-prompts/${row.id}`)
    } else {
      await api.delete(`/keys/${keyId}/mcp-prompts/${row.id}`)
    }
  } catch (e: any) {
    row.selected = previousValue
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function clearModels() {
  clearingModels.value = true
  try {
    await api.delete(`/keys/${keyId}/models`)
    ElMessage.success(t('common.success'))
    models.value.forEach(m => m.selected = false)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    clearingModels.value = false
  }
}

async function clearTools() {
  clearingTools.value = true
  try {
    await api.delete(`/keys/${keyId}/mcp-tools`)
    ElMessage.success(t('common.success'))
    tools.value.forEach(t => t.selected = false)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    clearingTools.value = false
  }
}

async function clearResources() {
  clearingResources.value = true
  try {
    await api.delete(`/keys/${keyId}/mcp-resources`)
    ElMessage.success(t('common.success'))
    resources.value.forEach(r => r.selected = false)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    clearingResources.value = false
  }
}

async function clearPrompts() {
  clearingPrompts.value = true
  try {
    await api.delete(`/keys/${keyId}/mcp-prompts`)
    ElMessage.success(t('common.success'))
    prompts.value.forEach(p => p.selected = false)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    clearingPrompts.value = false
  }
}

async function handleReset() {
  try {
    await ElMessageBox.confirm(
      t('key.resetConfirmMessage'),
      t('key.resetConfirmTitle'),
      { type: 'warning' }
    )
  } catch {
    return
  }

  resetting.value = true
  try {
    const res = await api.post(`/keys/${keyId}/reset`)
    key.value.key = res.data.key.key
    newKey.value = res.data.raw_key
    keyDialogVisible.value = true
    ElMessage.success(t('key.resetSuccess'))
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    resetting.value = false
  }
}

function copyKey() {
  navigator.clipboard.writeText(newKey.value)
  ElMessage.success(t('common.copied'))
}

function copyText(text: string) {
  if (!text) return
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success(t('common.copied'))
  }).catch(() => {
    ElMessage.error(t('common.error'))
  })
}

function goBack() {
  router.push('/keys')
}

function isLongText(text: string): boolean {
  if (!text) return false
  const lines = text.split('\n')
  return lines.length > 5 || text.length > 300
}

function handleSortChange(key: string, { prop, order }: any) {
  if (prop && order) {
    setSortConfig(key, { prop, order })
  }
}
</script>

<style scoped>
.key-detail {
  padding: 20px;
}

.info-card {
  margin-top: 20px;
}

.info-card :deep(.el-descriptions__table) {
  width: 100%;
  table-layout: fixed;
}

.info-card :deep(.el-descriptions__cell) {
  width: 50%;
}

.actions {
  margin-top: 20px;
  display: flex;
  gap: 10px;
}

.tabs-card {
  margin-top: 20px;
}

.tab-actions {
  margin-bottom: 16px;
  text-align: right;
}

code {
  background-color: var(--el-fill-color-light);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}

.description-cell {
  width: 100%;
}

.description-text {
  max-height: 7.5em;
  overflow: hidden;
  line-height: 1.5em;
  word-break: break-word;
  white-space: pre-wrap;
  position: relative;
}

.description-text.expanded {
  max-height: none;
}

.description-actions {
  margin-top: 4px;
  display: flex;
  gap: 8px;
}
</style>