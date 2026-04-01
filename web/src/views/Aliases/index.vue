<template>
  <div class="aliases-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('menu.aliases') }}</span>
          <el-button type="primary" @click="showAliasDialog()">{{ t('modelAlias.create') }}</el-button>
        </div>
      </template>

      <el-collapse v-model="activeNames" v-loading="loading">
        <el-collapse-item v-for="alias in aliases" :key="alias.id" :name="alias.id">
          <template #title>
            <div class="collapse-title">
              <span class="alias-name">{{ alias.name }}</span>
              <el-tag :type="alias.enabled ? 'success' : 'info'" size="small" style="margin-left: 8px">
                {{ alias.enabled ? t('common.enabled') : t('common.disabled') }}
              </el-tag>
              <span class="mapping-count">{{ alias.mapping_count }} {{ t('modelAlias.mappings') }}</span>
            </div>
          </template>
          <div class="mapping-actions">
            <el-button type="success" size="small" @click="showMappingDialog(alias.id)">{{ t('modelAlias.addMapping') }}</el-button>
            <el-button type="primary" size="small" @click="showAliasDialog(alias)">{{ t('common.edit') }}</el-button>
            <el-button type="danger" size="small" @click="handleDeleteAlias(alias.id)">{{ t('common.delete') }}</el-button>
          </div>
          <el-table :data="alias.mappings" stripe size="small">
            <el-table-column :label="t('provider.name')">
              <template #default="{ row }">{{ row.provider?.name }}</template>
            </el-table-column>
            <el-table-column :label="t('modelAlias.providerType')">
              <template #default="{ row }">
                <el-tag v-if="row.provider?.openai_base_url" type="success" size="small" style="margin-right: 4px">OpenAI</el-tag>
                <el-tag v-if="row.provider?.anthropic_base_url" type="primary" size="small">Anthropic</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="provider_model_name" :label="t('modelMapping.actualModel')" />
            <el-table-column prop="weight" :label="t('modelMapping.weight')" width="80" />
            <el-table-column :label="t('common.status')" width="80">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" @change="toggleMappingEnabled(alias.id, row)" />
              </template>
            </el-table-column>
            <el-table-column :label="t('common.action')" width="120">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="showMappingDialog(alias.id, row)">{{ t('common.edit') }}</el-button>
                <el-button link type="danger" size="small" @click="handleDeleteMapping(alias.id, row.id)">{{ t('common.delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-collapse-item>
      </el-collapse>

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

    <el-dialog v-model="mappingDialogVisible" :title="editingMapping ? t('modelAlias.editMapping') : t('modelAlias.addMapping')" width="500px">
      <el-form :model="mappingForm" :rules="mappingRules" ref="mappingFormRef" label-width="auto" v-loading="mappingDialogLoading">
        <el-form-item :label="t('provider.name')" prop="provider_id">
          <el-select v-model="mappingForm.provider_id" @change="loadProviderModels" style="width: 100%" filterable>
            <el-option v-for="p in providers" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('modelMapping.model')" prop="provider_model_name">
          <el-select v-model="mappingForm.provider_model_name" style="width: 100%" filterable :placeholder="mappingForm.provider_id ? '' : t('provider.name')">
            <el-option v-for="m in providerModels" :key="m.model_id" :label="m.model_id" :value="m.model_id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('modelMapping.weight')">
          <el-input-number v-model="mappingForm.weight" :min="1" :max="100" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-switch v-model="mappingForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mappingDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleMappingSubmit" :loading="submitting">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api'

const { t } = useI18n()

interface Mapping {
  id: number
  provider_id: number
  provider_model_name: string
  weight: number
  enabled: boolean
  provider?: {
    id: number
    name: string
    openai_base_url: string
    anthropic_base_url: string
  }
}

interface Alias {
  id: number
  name: string
  enabled: boolean
  mapping_count: number
  mappings: Mapping[]
}

const aliases = ref<Alias[]>([])
const providers = ref<any[]>([])
const providerModels = ref<any[]>([])
const activeNames = ref<number[]>([])
const loading = ref(false)
const aliasDialogVisible = ref(false)
const mappingDialogVisible = ref(false)
const mappingDialogLoading = ref(false)
const submitting = ref(false)
const editingAlias = ref<Alias | null>(null)
const editingMapping = ref<Mapping | null>(null)
const currentAliasId = ref<number | null>(null)
const aliasFormRef = ref()
const mappingFormRef = ref()

const aliasForm = reactive({
  name: '',
  enabled: true
})

const mappingForm = reactive({
  provider_id: null as number | null,
  provider_model_name: '',
  weight: 1,
  enabled: true
})

const aliasRules = computed(() => ({
  name: [{ required: true, message: t('common.required'), trigger: 'blur' }]
}))

const mappingRules = computed(() => ({
  provider_id: [{ required: true, message: t('common.required'), trigger: 'change' }],
  provider_model_name: [{ required: true, message: t('common.required'), trigger: 'change' }]
}))

onMounted(() => {
  fetchAliases()
  fetchProviders()
})

async function fetchAliases() {
  loading.value = true
  try {
    const res = await api.get('/aliases')
    const list = res.data.aliases || []
    aliases.value = await Promise.all(list.map(async (a: any) => {
      const detailRes = await api.get(`/aliases/${a.id}`)
      return {
        id: a.id,
        name: a.alias,
        enabled: a.enabled,
        mapping_count: a.mapping_count,
        mappings: detailRes.data.alias?.mappings || []
      }
    }))
  } finally {
    loading.value = false
  }
}

async function fetchProviders() {
  const res = await api.get('/providers')
  providers.value = (res.data.providers || []).sort((a: any, b: any) => a.name.localeCompare(b.name))
}

async function loadProviderModels() {
  if (!mappingForm.provider_id) return
  const res = await api.get(`/providers/${mappingForm.provider_id}/models`)
  providerModels.value = (res.data.models || []).sort((a: any, b: any) => a.model_id.localeCompare(b.model_id))
}

function showAliasDialog(alias?: Alias) {
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

function showMappingDialog(aliasId: number, mapping?: Mapping) {
  currentAliasId.value = aliasId
  editingMapping.value = mapping || null
  Object.assign(mappingForm, {
    provider_id: mapping?.provider_id || null,
    provider_model_name: mapping?.provider_model_name || '',
    weight: mapping?.weight || 1,
    enabled: mapping?.enabled ?? true
  })
  providerModels.value = []
  if (mappingForm.provider_id) {
    loadProviderModels()
  }
  mappingDialogVisible.value = true
}

async function handleMappingSubmit() {
  const valid = await mappingFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingMapping.value) {
      await api.put(`/aliases/${currentAliasId.value}/mappings/${editingMapping.value.id}`, mappingForm)
    } else {
      await api.post(`/aliases/${currentAliasId.value}/mappings`, mappingForm)
    }
    ElMessage.success(t('common.success'))
    mappingDialogVisible.value = false
    fetchAliases()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    submitting.value = false
  }
}

async function handleDeleteMapping(aliasId: number, mappingId: number) {
  await ElMessageBox.confirm(t('common.confirm'), t('common.delete'), { type: 'warning' })
  await api.delete(`/aliases/${aliasId}/mappings/${mappingId}`)
  ElMessage.success(t('common.success'))
  fetchAliases()
}

async function toggleMappingEnabled(aliasId: number, mapping: Mapping) {
  await api.put(`/aliases/${aliasId}/mappings/${mapping.id}`, { enabled: mapping.enabled })
}
</script>

<style scoped>
.aliases-page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.collapse-title { display: flex; align-items: center; width: 100%; }
.alias-name { font-weight: 500; }
.mapping-count { margin-left: auto; color: #909399; font-size: 13px; }
.mapping-actions { margin-bottom: 12px; display: flex; gap: 8px; }
</style>