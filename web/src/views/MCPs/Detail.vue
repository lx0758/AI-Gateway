<template>
  <div class="mcp-service-detail">
    <el-page-header @back="goBack" :title="t('mcp.services')">
      <template #content>
        <span class="text-large font-600 mr-3">{{ service?.name || '-' }}</span>
      </template>
    </el-page-header>

    <el-card class="info-card" v-loading="loading">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="t('mcp.name')">{{ service?.name }}</el-descriptions-item>
        <el-descriptions-item :label="t('mcp.type')">
          <el-tag :type="service?.type === 'remote' ? 'success' : 'info'">{{ service?.type }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('mcp.lastSync')">
          {{ service?.last_sync_at ? new Date(service.last_sync_at).toLocaleString() : '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('mcp.status')">
          <el-tag :type="service?.enabled ? 'success' : 'info'" size="small">
            {{ service?.enabled ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="service?.type === 'remote' ? t('mcp.url') : t('mcp.command')" :span="2">
          <code v-if="service?.type === 'local'">{{ service?.target }}</code>
          <span v-else>{{ service?.target }}</span>
        </el-descriptions-item>
      </el-descriptions>
      <div class="actions">
        <el-button @click="testConnection" :loading="testing">{{ t('mcp.testConnection') }}</el-button>
        <el-button type="primary" @click="syncService" :loading="syncing">{{ t('mcp.sync') }}</el-button>
      </div>
    </el-card>

    <el-card class="tabs-card">
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="t('mcp.tools')" name="tools">
          <el-table :data="tools" stripe v-loading="toolsLoading">
            <el-table-column prop="name" :label="t('mcp.toolName')" width="200" />
            <el-table-column :label="t('mcp.toolDescription')">
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
            <el-table-column :label="t('common.status')" width="100">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" @change="toggleToolEnabled(row)" />
              </template>
            </el-table-column>
            <el-table-column :label="t('common.action')" width="150">
              <template #default="{ row }">
                <el-button link type="primary" @click="showToolDetail(row)">{{ t('common.detail') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('mcp.resources')" name="resources">
          <el-table :data="resources" stripe v-loading="resourcesLoading">
            <el-table-column prop="uri" :label="t('mcp.resourceUri')" width="300" />
            <el-table-column prop="name" :label="t('mcp.resourceName')" width="200" />
            <el-table-column prop="mime_type" label="MIME Type" width="150" />
            <el-table-column :label="t('common.status')" width="100">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" @change="toggleResourceEnabled(row)" />
              </template>
            </el-table-column>
            <el-table-column :label="t('common.action')" width="150">
              <template #default="{ row }">
                <el-button link type="primary" @click="showResourceDetail(row)">{{ t('common.detail') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('mcp.prompts')" name="prompts">
          <el-table :data="prompts" stripe v-loading="promptsLoading">
            <el-table-column prop="name" :label="t('mcp.promptName')" width="200" />
            <el-table-column :label="t('mcp.toolDescription')">
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
            <el-table-column :label="t('common.status')" width="100">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" @change="togglePromptEnabled(row)" />
              </template>
            </el-table-column>
            <el-table-column :label="t('common.action')" width="150">
              <template #default="{ row }">
                <el-button link type="primary" @click="showPromptDetail(row)">{{ t('common.detail') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="detailDialogVisible" :title="detailTitle" width="85%" class="mcp-detail-dialog" :style="{ margin: '5vh auto' }">
      <el-descriptions :column="1" border>
        <el-descriptions-item :label="t('common.name')">{{ detailData.name }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.description')">
          <div class="description-cell">
            <div class="description-text" :class="{ expanded: detailData._descExpanded }">
              {{ detailData.description || '-' }}
            </div>
            <div class="description-actions">
              <el-button v-if="detailData.description && isLongText(detailData.description)" link type="primary" size="small" @click="detailData._descExpanded = !detailData._descExpanded">
                {{ detailData._descExpanded ? t('common.collapse') : t('common.expand') }}
              </el-button>
              <el-button v-if="detailData.description" link type="primary" size="small" @click="copyText(detailData.description)">
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
          </div>
        </el-descriptions-item>
        <el-descriptions-item v-if="detailData.mime_type" label="MIME Type">{{ detailData.mime_type }}</el-descriptions-item>
        <el-descriptions-item v-if="detailData.uri" :label="t('mcp.resourceUri')">
          <code>{{ detailData.uri }}</code>
        </el-descriptions-item>
      </el-descriptions>
      
      <div v-if="detailData.input_schema || detailData.arguments" style="margin-top: 16px">
        <JsonViewer :json="detailData.input_schema || detailData.arguments" :title="detailData.input_schema ? t('mcp.inputSchema') : 'Arguments'" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import JsonViewer from '@/components/JsonViewer.vue'
import api from '@/api'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const serviceId = Number(route.params.id)
const service = ref<any>(null)
const loading = ref(false)
const syncing = ref(false)
const testing = ref(false)
const activeTab = ref('tools')

const tools = ref<any[]>([])
const resources = ref<any[]>([])
const prompts = ref<any[]>([])
const toolsLoading = ref(false)
const resourcesLoading = ref(false)
const promptsLoading = ref(false)

const detailDialogVisible = ref(false)
const detailTitle = ref('')
const detailData = ref<any>({})

onMounted(() => {
  fetchService()
})

watch(activeTab, (newTab) => {
  if (newTab === 'tools' && tools.value.length === 0) fetchTools()
  if (newTab === 'resources' && resources.value.length === 0) fetchResources()
  if (newTab === 'prompts' && prompts.value.length === 0) fetchPrompts()
})

async function fetchService() {
  loading.value = true
  try {
    const res = await api.get(`/mcps/${serviceId}`)
    service.value = res.data.mcp
    fetchTools()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
    goBack()
  } finally {
    loading.value = false
  }
}

async function fetchTools() {
  toolsLoading.value = true
  try {
    const res = await api.get(`/mcps/${serviceId}/tools`)
    tools.value = res.data.tools || []
  } finally {
    toolsLoading.value = false
  }
}

async function fetchResources() {
  resourcesLoading.value = true
  try {
    const res = await api.get(`/mcps/${serviceId}/resources`)
    resources.value = res.data.resources || []
  } finally {
    resourcesLoading.value = false
  }
}

async function fetchPrompts() {
  promptsLoading.value = true
  try {
    const res = await api.get(`/mcps/${serviceId}/prompts`)
    prompts.value = res.data.prompts || []
  } finally {
    promptsLoading.value = false
  }
}

async function syncService() {
  syncing.value = true
  try {
    const res = await api.post(`/mcps/${serviceId}/sync`)
    if (res.data.success) {
      ElMessage.success(t('mcp.syncSuccess'))
      fetchService()
      if (activeTab.value === 'tools') fetchTools()
      if (activeTab.value === 'resources') fetchResources()
      if (activeTab.value === 'prompts') fetchPrompts()
    } else {
      ElMessage.error(res.data.message || t('mcp.syncFailed'))
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('mcp.syncFailed'))
  } finally {
    syncing.value = false
  }
}

async function testConnection() {
  testing.value = true
  try {
    const res = await api.post(`/mcps/${serviceId}/test`)
    if (res.data.success) {
      ElMessage.success(t('mcp.connectionSuccess'))
    } else {
      ElMessage.error(res.data.message || t('mcp.connectionFailed'))
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('mcp.connectionFailed'))
  } finally {
    testing.value = false
  }
}

function goBack() {
  router.push('/mcps')
}

function goEdit() {
  router.push(`/mcps/${serviceId}/edit`)
}

function showToolDetail(tool: any) {
  detailTitle.value = `${t('mcp.tools')}: ${tool.name}`
  detailData.value = tool
  detailDialogVisible.value = true
}

function showResourceDetail(resource: any) {
  detailTitle.value = `${t('mcp.resources')}: ${resource.name}`
  detailData.value = resource
  detailDialogVisible.value = true
}

function showPromptDetail(prompt: any) {
  detailTitle.value = `${t('mcp.prompts')}: ${prompt.name}`
  detailData.value = prompt
  detailDialogVisible.value = true
}

function isLongText(text: string): boolean {
  if (!text) return false
  const lines = text.split('\n')
  return lines.length > 5 || text.length > 300
}

function copyText(text: string) {
  if (!text) return
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success(t('common.copied'))
  }).catch(() => {
    ElMessage.error(t('common.error'))
  })
}

async function toggleToolEnabled(tool: any) {
  try {
    await api.put(`/mcps/tools/${tool.id}`, { enabled: tool.enabled })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    tool.enabled = !tool.enabled
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function toggleResourceEnabled(resource: any) {
  try {
    await api.put(`/mcps/resources/${resource.id}`, { enabled: resource.enabled })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    resource.enabled = !resource.enabled
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}

async function togglePromptEnabled(prompt: any) {
  try {
    await api.put(`/mcps/prompts/${prompt.id}`, { enabled: prompt.enabled })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    prompt.enabled = !prompt.enabled
    ElMessage.error(e.response?.data?.error || t('common.error'))
  }
}
</script>

<style scoped>
.mcp-service-detail {
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

.extra-actions {
  display: flex;
  gap: 10px;
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

<style>
.el-dialog.mcp-detail-dialog .el-dialog__header {
  padding: 12px 16px;
  margin: 0;
}

.el-dialog.mcp-detail-dialog .el-dialog__body {
  padding: 12px 16px;
  padding-top: 0;
}

.el-dialog.mcp-detail-dialog .el-dialog__footer {
  padding: 12px 16px;
}

.el-dialog.mcp-detail-dialog .el-descriptions {
  margin: 0;
}
</style>
