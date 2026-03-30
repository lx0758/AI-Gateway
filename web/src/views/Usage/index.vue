<template>
  <div class="usage-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('usage.stats') }}</span>
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            :range-separator="t('common.to')"
            :start-placeholder="t('usage.startTime')"
            :end-placeholder="t('usage.endTime')"
            value-format="YYYY-MM-DD HH:mm:ss"
            @change="fetchStats"
          />
        </div>
      </template>
      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic :title="t('usage.totalRequests')" :value="stats.totalRequests" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('usage.successRate')" :value="stats.successRate?.toFixed(1)" suffix="%" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('usage.totalTokens')" :value="stats.totalTokens" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('usage.promptTokens')" :value="stats.promptTokens" />
        </el-col>
      </el-row>
    </el-card>

    <el-card class="model-stats-card" v-if="stats.modelStats?.length">
      <template #header>{{ t('usage.modelStats') || '模型统计' }}</template>
      <el-table :data="stats.modelStats" stripe size="small">
        <el-table-column prop="model" :label="t('usage.model') || '映射模型'" />
        <el-table-column prop="actual_model" :label="t('usage.actualModel') || '实际模型'" />
        <el-table-column prop="count" :label="t('usage.callCount') || '调用次数'" width="120" />
        <el-table-column prop="tokens" :label="t('usage.totalTokens') || 'Token'" width="120" />
      </el-table>
    </el-card>

    <el-card class="logs-card">
      <template #header>{{ t('usage.logs') }}</template>
      <el-table :data="logs" stripe v-loading="loading">
        <el-table-column :label="t('usage.time')" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="model" :label="t('usage.model')">
          <template #default="{ row }">
            <span>{{ row.model }}</span>
            <el-tag v-if="row.actual_model && row.actual_model !== row.model" size="small" type="info" style="margin-left: 4px">
              {{ row.actual_model }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="prompt_tokens" :label="t('usage.promptTokens')" width="100" />
        <el-table-column prop="completion_tokens" :label="t('usage.completionTokens')" width="120" />
        <el-table-column prop="latency_ms" :label="t('usage.avgLatency')" width="100">
          <template #default="{ row }">{{ row.latency_ms }}ms</template>
        </el-table-column>
        <el-table-column prop="status" :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

const stats = ref<any>({
  totalRequests: 0,
  successRate: 0,
  totalTokens: 0,
  promptTokens: 0,
  modelStats: []
})
const logs = ref<any[]>([])
const loading = ref(false)
const dateRange = ref<string[] | null>(null)

onMounted(() => {
  fetchStats()
  fetchLogs()
})

async function fetchStats() {
  try {
    const res = await api.get('/usage/stats')
    stats.value = res.data
  } catch (e) {
    console.error(e)
  }
}

async function fetchLogs() {
  loading.value = true
  try {
    const res = await api.get('/usage/logs')
    logs.value = res.data.logs || []
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.usage-page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.model-stats-card { margin-top: 20px; }
.logs-card { margin-top: 20px; }
</style>
