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
          <el-statistic :title="t('usage.totalTokens') || '总 Tokens'" :value="stats.totalTokens" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('usage.avgLatency') || '平均耗时'" :value="formatLatency(stats.avgLatency)" />
        </el-col>
      </el-row>
    </el-card>

    <el-card class="key-stats-card" v-if="keyStats.length">
      <template #header>{{ t('usage.keyStats') || 'Key 统计' }}</template>
       <el-table :data="keyStats" stripe size="small">
        <el-table-column prop="key_name" :label="t('usage.keyName') || 'Key 名称'" />
        <el-table-column prop="count" :label="t('usage.callCount') || '调用次数'" />
        <el-table-column prop="tokens" :label="'Tokens'">
          <template #default="{ row }">{{ formatTokens(row.tokens) }}</template>
        </el-table-column>
        <el-table-column :label="t('usage.avgLatency') || '平均耗时'">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="model-stats-card" v-if="stats.modelStats?.length">
      <template #header>{{ t('usage.modelStats') || '模型统计' }}</template>
      <el-table :data="stats.modelStats" stripe size="small">
        <el-table-column prop="model" :label="t('usage.model') || '映射模型'" width="200" />
        <el-table-column prop="actual_model" :label="t('usage.providerModel') || '厂家/模型'" width="300" >
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.provider_name }}/{{ row.actual_model }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="count" :label="t('usage.callCount') || '调用次数'" />
        <el-table-column prop="tokens" :label="'Tokens'">
          <template #default="{ row }">{{ formatTokens(row.tokens) }}</template>
        </el-table-column>
        <el-table-column :label="t('usage.avgLatency') || '平均耗时'">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="logs-card">
      <template #header>{{ t('usage.logs') }}</template>
      <el-table :data="logs" stripe v-loading="loading" size="small">
        <el-table-column :label="t('usage.time') || '时间'" width="160">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="source" :label="t('usage.source') || '来源'" width="120">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.source }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="key_name" :label="t('usage.key') || 'Key'" width="120" />
        <el-table-column prop="model" :label="t('usage.model') || '模型'" width="150">
          <template #default="{ row }">
            <span>{{ row.model }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('usage.providerModel') || '厂家/模型'" width="280">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.provider_name }}/{{ row.actual_model }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_tokens" :label="'Tokens'" width="100">
          <template #default="{ row }">{{ formatTokens(row.total_tokens) }}</template>
        </el-table-column>
        <el-table-column prop="latency_ms" :label="t('usage.latency') || '耗时'" width="100">
          <template #default="{ row }">{{ formatLatency(row.latency_ms) }}</template>
        </el-table-column>
        <el-table-column prop="status" :label="t('common.status') || '结果'" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="error_msg" :label="t('usage.error') || '错误信息'" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.error_msg" class="error-text">{{ row.error_msg }}</span>
            <span v-else>-</span>
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
import { formatDateTime, formatLatency, formatTokens } from '@/utils/format'

const { t } = useI18n()

const stats = ref<any>({
  totalRequests: 0,
  successRate: 0,
  totalTokens: 0,
  avgLatency: 0,
  modelStats: []
})
const keyStats = ref<any[]>([])
const logs = ref<any[]>([])
const loading = ref(false)
const dateRange = ref<string[] | null>(null)

onMounted(() => {
  fetchStats()
  fetchKeyStats()
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

async function fetchKeyStats() {
  try {
    const res = await api.get('/usage/key-stats')
    keyStats.value = res.data.keyStats || []
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
.key-stats-card { margin-top: 20px; }
.model-stats-card { margin-top: 20px; }
.logs-card { margin-top: 20px; }
.error-text { 
  color: var(--el-color-danger); 
  font-size: 12px;
}
</style>
