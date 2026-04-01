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
            @change="fetchLogs"
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
          <el-statistic :title="t('usage.totalTokens') || '总 Tokens'" :value="formatTokens(stats.totalTokens)" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('usage.avgLatency') || '平均耗时'" :value="formatLatency(stats.avgLatency)" />
        </el-col>
      </el-row>
    </el-card>

    <el-card class="source-stats-card" v-if="sourceStats.length">
      <template #header>{{ t('usage.sourceStats') || '接入点统计' }}</template>
      <el-table :data="sourceStats" stripe size="small">
        <el-table-column prop="source" :label="t('usage.source') || '接入点'" />
        <el-table-column prop="count" :label="t('usage.callCount') || '调用次数'" />
        <el-table-column prop="tokens" :label="'Tokens'">
          <template #default="{ row }">{{ formatTokens(row.tokens) }}</template>
        </el-table-column>
        <el-table-column :label="t('usage.avgLatency') || '平均耗时'">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="call-method-stats-card" v-if="callMethodStats.length">
      <template #header>{{ t('usage.callMethodStats') || '调用方式统计' }}</template>
      <el-table :data="callMethodStats" stripe size="small">
        <el-table-column prop="call_method" :label="t('usage.callMethod') || '调用方式'" />
        <el-table-column prop="count" :label="t('usage.callCount') || '调用次数'" />
        <el-table-column prop="tokens" :label="'Tokens'">
          <template #default="{ row }">{{ formatTokens(row.tokens) }}</template>
        </el-table-column>
        <el-table-column :label="t('usage.avgLatency') || '平均耗时'">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
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

    <el-card class="model-stats-card" v-if="modelStats.length">
      <template #header>{{ t('usage.modelStats') || '模型统计' }}</template>
      <el-table :data="modelStats" stripe size="small">
        <el-table-column prop="model" :label="t('usage.model') || '模型'" />
        <el-table-column prop="count" :label="t('usage.callCount') || '调用次数'" />
        <el-table-column prop="tokens" :label="'Tokens'">
          <template #default="{ row }">{{ formatTokens(row.tokens) }}</template>
        </el-table-column>
        <el-table-column :label="t('usage.avgLatency') || '平均耗时'">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="provider-stats-card" v-if="providerStats.length">
      <template #header>{{ t('usage.providerStats') || '厂商统计' }}</template>
      <el-table :data="providerStats" stripe size="small">
        <el-table-column prop="provider_name" :label="t('usage.provider') || '厂商'" />
        <el-table-column prop="count" :label="t('usage.callCount') || '调用次数'" />
        <el-table-column prop="tokens" :label="'Tokens'">
          <template #default="{ row }">{{ formatTokens(row.tokens) }}</template>
        </el-table-column>
        <el-table-column :label="t('usage.avgLatency') || '平均耗时'">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="provider-model-stats-card" v-if="providerModelStats.length">
      <template #header>{{ t('usage.providerModelStats') || '厂商模型统计' }}</template>
      <el-table :data="providerModelStats" stripe size="small">
        <el-table-column :label="t('usage.model')">
          <template #default="{ row }">
            {{ row.actual_model_name }}
          </template>
        </el-table-column>
        <el-table-column :label="t('usage.provider')">
          <template #default="{ row }">
            {{ row.provider_name }}
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
        <el-table-column prop="source" :label="t('usage.source') || '来源'" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.source }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="key_name" :label="t('usage.key') || 'Key'" width="100" />
        <el-table-column prop="model" :label="t('usage.model') || '模型'" width="150">
          <template #default="{ row }">
            <span>{{ row.model }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('usage.providerModel') || '厂商/模型'" width="250">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.provider_name }}/{{ row.actual_model_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="call_method" :label="t('usage.callMethod') || '调用方式'" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.call_method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_tokens" :label="'Tokens'" width="100">
          <template #default="{ row }">{{ formatTokens(row.total_tokens) }}</template>
        </el-table-column>
        <el-table-column prop="latency_ms" :label="t('usage.latency') || '耗时'" width="80">
          <template #default="{ row }">{{ formatLatency(row.latency_ms) }}</template>
        </el-table-column>
        <el-table-column prop="status" :label="t('common.status') || '状态'" width="100">
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
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import api from '@/api'
import { formatDateTime, formatLatency, formatTokens } from '@/utils/format'

const { t } = useI18n()

interface LogItem {
  id: number
  source: string
  key_id: number
  key_name: string
  model: string
  provider_id: number
  provider_name: string
  actual_model_id: string
  actual_model_name: string
  call_method: string
  total_tokens: number
  latency_ms: number
  status: string
  error_msg: string
  created_at: string
}

const logs = ref<LogItem[]>([])
const loading = ref(false)
const dateRange = ref<string[] | null>(null)

const stats = computed(() => {
  const list = logs.value
  if (list.length === 0) {
    return { totalRequests: 0, successRate: 0, totalTokens: 0, avgLatency: 0 }
  }
  const totalRequests = list.length
  const successCount = list.filter(l => l.status === 'success').length
  const successRate = (successCount / totalRequests) * 100
  const totalTokens = list.reduce((sum, l) => sum + (l.total_tokens || 0), 0)
  const avgLatency = list.reduce((sum, l) => sum + (l.latency_ms || 0), 0) / list.length
  return { totalRequests, successRate, totalTokens, avgLatency }
})

const sourceStats = computed(() => aggregateBy('source'))
const callMethodStats = computed(() => aggregateBy('call_method'))
const keyStats = computed(() => aggregateBy('key_name'))
const modelStats = computed(() => aggregateBy('model'))
const providerStats = computed(() => aggregateBy('provider_name'))
const providerModelStats = computed(() => aggregateBy(['provider_name', 'actual_model_name']))

function aggregateBy(dimensions: string | string[]): any[] {
  const list = logs.value
  const dimKey = Array.isArray(dimensions) ? dimensions.join('_') : dimensions
  const groups: Record<string, { count: number; tokens: number; latency: number }> = {}

  for (const log of list) {
    let key: string
    if (Array.isArray(dimensions)) {
      const values = dimensions.map(d => (log as any)[d] || 'unknown')
      key = values.join('_')
    } else {
      key = (log as any)[dimensions] || 'unknown'
    }

    if (!groups[key]) {
      groups[key] = { count: 0, tokens: 0, latency: 0 }
    }
    groups[key].count++
    groups[key].tokens += log.total_tokens || 0
    groups[key].latency += log.latency_ms || 0
  }

  return Object.entries(groups)
    .map(([key, value]) => {
      const item: any = {
        count: value.count,
        tokens: value.tokens,
        avg_latency: value.count > 0 ? value.latency / value.count : 0
      }
      if (Array.isArray(dimensions)) {
        dimensions.forEach((d, i) => {
          item[d] = key.split('_')[i]
        })
      } else {
        item[dimensions] = key
      }
      return item
    })
    .sort((a, b) => b.count - a.count)
}

onMounted(() => {
  fetchLogs()
})

async function fetchLogs() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0].split(' ')[0]
      params.end_date = dateRange.value[1].split(' ')[0]
    }
    const res = await api.get('/usage/logs', { params })
    logs.value = res.data.logs || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.usage-page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.source-stats-card { margin-top: 20px; }
.call-method-stats-card { margin-top: 20px; }
.key-stats-card { margin-top: 20px; }
.model-stats-card { margin-top: 20px; }
.provider-stats-card { margin-top: 20px; }
.provider-model-stats-card { margin-top: 20px; }
.logs-card { margin-top: 20px; }
.error-text { 
  color: var(--el-color-danger); 
  font-size: 12px;
}
</style>
