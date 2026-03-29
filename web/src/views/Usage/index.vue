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
          <el-statistic :title="t('usage.totalRequests')" :value="stats.total_requests" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('usage.successRate')" :value="stats.success_rate" suffix="%" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('usage.totalTokens')" :value="stats.total_tokens" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('usage.promptTokens')" :value="stats.prompt_tokens" />
        </el-col>
      </el-row>
    </el-card>

    <el-card class="logs-card">
      <template #header>{{ t('usage.logs') }}</template>
      <el-table :data="logs" stripe v-loading="loading">
        <el-table-column :label="t('usage.time')" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="model" :label="t('usage.model')" />
        <el-table-column prop="prompt_tokens" :label="t('usage.promptTokens')" />
        <el-table-column prop="completion_tokens" :label="t('usage.completionTokens')" />
        <el-table-column prop="latency_ms" :label="t('usage.avgLatency')">
          <template #default="{ row }">{{ row.latency_ms }}ms</template>
        </el-table-column>
        <el-table-column prop="status" :label="t('common.status')">
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
  total_requests: 0,
  success_rate: 0,
  total_tokens: 0,
  prompt_tokens: 0
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
.logs-card { margin-top: 20px; }
</style>
