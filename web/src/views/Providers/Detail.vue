<template>
  <div class="provider-detail">
    <el-page-header @back="$router.back()">
      <template #content>
        {{ provider?.name || t('menu.providers') }}
      </template>
    </el-page-header>

    <el-card class="info-card">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="t('provider.apiType')">{{ provider?.api_type }}</el-descriptions-item>
        <el-descriptions-item :label="t('provider.baseUrl')">{{ provider?.base_url }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">
          <el-tag :type="provider?.enabled ? 'success' : 'info'">
            {{ provider?.enabled ? t('common.enabled') : t('common.disabled') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('provider.lastSync')">{{ formatDateTime(provider?.last_sync_at) }}</el-descriptions-item>
      </el-descriptions>
      <div class="actions">
        <el-button type="primary" @click="syncModels" :loading="syncing">{{ t('provider.syncModels') }}</el-button>
      </div>
    </el-card>

    <el-card>
      <template #header>{{ t('provider.models') }}</template>
      <el-table :data="models" stripe>
        <el-table-column prop="model_id" :label="t('provider.modelId')" />
        <el-table-column prop="display_name" :label="t('common.name')" />
        <el-table-column prop="context_window" :label="t('provider.contextWindow')" />
        <el-table-column :label="t('common.status')">
          <template #default="{ row }">
            <el-tag :type="row.is_available ? 'success' : 'danger'">
              {{ row.is_available ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import api from '@/api'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const route = useRoute()

const provider = ref<any>(null)
const models = ref<any[]>([])
const syncing = ref(false)

const providerId = route.params.id as string

onMounted(() => {
  fetchProvider()
})

async function fetchProvider() {
  try {
    const res = await api.get(`/providers/${providerId}`)
    provider.value = res.data.provider
    models.value = res.data.provider?.models || []
  } catch (e) {
    console.error(e)
  }
}

async function syncModels() {
  syncing.value = true
  try {
    await api.post(`/providers/${providerId}/sync`)
    ElMessage.success(t('common.success'))
    fetchProvider()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    syncing.value = false
  }
}
</script>

<style scoped>
.provider-detail {
  padding: 20px;
}

.info-card {
  margin: 20px 0;
}

.actions {
  margin-top: 20px;
}
</style>
