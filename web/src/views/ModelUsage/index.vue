<template>
  <div class="usage-page">
    <el-card v-loading="loading">
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
            style="width: 500px; flex: 0 0 500px"
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
          <template #default="{ row }">C:{{ formatTokens(row.cached_tokens) }}/I:{{ formatTokens(row.input_tokens) }}/O:{{ formatTokens(row.output_tokens) }}/T:{{ formatTokens(row.total_tokens) }}</template>
        </el-table-column>
        <el-table-column :label="t('usage.avgLatency') || '平均耗时'">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="ip-stats-card" v-if="ipStats.length">
      <template #header>{{ t('usage.ipStats') || 'IP 统计' }}</template>
      <el-table :data="ipStats" stripe size="small">
        <el-table-column :label="t('usage.clientIp') || '客户端 IP'">
          <template #default="{ row }">
            <span>{{ row.client_ips }}</span>
            <el-tooltip v-if="row.full_chain.includes(',')" :content="row.full_chain" placement="top">
              <el-icon style="margin-left: 4px; cursor: help;"><InfoFilled /></el-icon>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="count" :label="t('usage.callCount') || '调用次数'" />
        <el-table-column :label="'Tokens'">
          <template #default="{ row }">C:{{ formatTokens(row.cached_tokens) }}/I:{{ formatTokens(row.input_tokens) }}/O:{{ formatTokens(row.output_tokens) }}/T:{{ formatTokens(row.total_tokens) }}</template>
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
        <el-table-column :label="'Tokens'">
          <template #default="{ row }">C:{{ formatTokens(row.cached_tokens) }}/I:{{ formatTokens(row.input_tokens) }}/O:{{ formatTokens(row.output_tokens) }}/T:{{ formatTokens(row.total_tokens) }}</template>
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
        <el-table-column :label="'Tokens'">
          <template #default="{ row }">C:{{ formatTokens(row.cached_tokens) }}/I:{{ formatTokens(row.input_tokens) }}/O:{{ formatTokens(row.output_tokens) }}/T:{{ formatTokens(row.total_tokens) }}</template>
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
        <el-table-column :label="'Tokens'">
          <template #default="{ row }">C:{{ formatTokens(row.cached_tokens) }}/I:{{ formatTokens(row.input_tokens) }}/O:{{ formatTokens(row.output_tokens) }}/T:{{ formatTokens(row.total_tokens) }}</template>
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
        <el-table-column :label="'Tokens'">
          <template #default="{ row }">C:{{ formatTokens(row.cached_tokens) }}/I:{{ formatTokens(row.input_tokens) }}/O:{{ formatTokens(row.output_tokens) }}/T:{{ formatTokens(row.total_tokens) }}</template>
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
        <el-table-column :label="'Tokens'">
          <template #default="{ row }">C:{{ formatTokens(row.cached_tokens) }}/I:{{ formatTokens(row.input_tokens) }}/O:{{ formatTokens(row.output_tokens) }}/T:{{ formatTokens(row.total_tokens) }}</template>
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
        <el-table-column prop="client_ips" :label="t('usage.clientIp') || 'IP'" width="120">
          <template #default="{ row }">
            <span>{{ row.client_ips.split(',')[0].trim() }}</span>
            <el-tooltip v-if="row.client_ips.includes(',')" :content="row.client_ips" placement="top">
              <el-icon style="margin-left: 4px; cursor: help;"><InfoFilled /></el-icon>
            </el-tooltip>
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
        <el-table-column :label="'Tokens'" width="200">
          <template #default="{ row }">C:{{ formatTokens(row.cached_tokens) }}/I:{{ formatTokens(row.input_tokens) }}/O:{{ formatTokens(row.output_tokens) }}/T:{{ formatTokens(row.total_tokens) }}</template>
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
            <div v-if="row.error_msg" class="error-cell">
              <span class="error-text">{{ row.error_msg }}</span>
              <el-button link type="primary" size="small" @click="copyError(row.error_msg)" style="margin-left: 4px">
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
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
import { ElMessage } from 'element-plus'
import { CopyDocument, InfoFilled } from '@element-plus/icons-vue'
import api from '@/api'
import { formatDateTime, formatLatency, formatTokens } from '@/utils/format'

const { t } = useI18n()

interface LogItem {
  id: number
  source: string
  client_ips: string
  key_id: number
  key_name: string
  model: string
  provider_id: number
  provider_name: string
  actual_model_id: string
  actual_model_name: string
  call_method: string
  cached_tokens: number
  input_tokens: number
  output_tokens: number
  total_tokens: number
  latency_ms: number
  status: string
  error_msg: string
  created_at: string
}

const logs = ref<LogItem[]>([])
const loading = ref(false)

function getDefaultDateRange(): string[] {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0)
  const end = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1, 0, 0, 0)
  const pad = (n: number) => n.toString().padStart(2, '0')
  const format = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  return [format(start), format(end)]
}

const dateRange = ref<string[]>(getDefaultDateRange())

const stats = computed(() => {
  const list = logs.value
  if (list.length === 0) {
    return {
      totalRequests: 0,
      successRate: 0,
      cachedTokens: 0,
      inputTokens: 0,
      outputTokens: 0,
      totalTokens: 0,
      avgLatency: 0,
    }
  }
  const totalRequests = list.length
  const successCount = list.filter(l => l.status === 'success').length
  const successRate = (successCount / totalRequests) * 100
  const cachedTokens = list.reduce((sum, l) => sum + (l.cached_tokens || 0), 0)
  const inputTokens = list.reduce((sum, l) => sum + (l.input_tokens || 0), 0)
  const outputTokens = list.reduce((sum, l) => sum + (l.output_tokens || 0), 0)
  const totalTokens = list.reduce((sum, l) => sum + (l.total_tokens || 0), 0)
  const avgLatency = list.reduce((sum, l) => sum + (l.latency_ms || 0), 0) / list.length
  return {
    totalRequests,
    successRate,
    cachedTokens,
    inputTokens,
    outputTokens,
    totalTokens,
    avgLatency
  }
})

const sourceStats = computed(() => aggregateBy('source'))
const callMethodStats = computed(() => aggregateBy('call_method'))
const keyStats = computed(() => aggregateBy('key_name'))
const ipStats = computed(() => {
  const list = logs.value
  const groups: Record<string, {
    count: number;
    cached_tokens: number;
    input_tokens: number;
    output_tokens: number;
    total_tokens: number;
    latency: number;
    full_chain: string;
  }> = {}

  for (const log of list) {
    const chain = log.client_ips || 'unknown'
    const firstIP = chain.split(',')[0].trim()
    if (!groups[firstIP]) {
      groups[firstIP] = {
        count: 0,
        cached_tokens: 0,
        input_tokens: 0,
        output_tokens: 0,
        total_tokens: 0,
        latency: 0,
        full_chain: chain,
      }
    }
    groups[firstIP].count++
    groups[firstIP].cached_tokens += log.cached_tokens || 0
    groups[firstIP].input_tokens += log.input_tokens || 0
    groups[firstIP].output_tokens += log.output_tokens || 0
    groups[firstIP].total_tokens += log.total_tokens || 0
    groups[firstIP].latency += log.latency_ms || 0
  }

  return Object.entries(groups)
    .map(([key, value]) => ({
      client_ips: key,
      full_chain: value.full_chain,
      count: value.count,
      cached_tokens: value.cached_tokens,
      input_tokens: value.input_tokens,
      output_tokens: value.output_tokens,
      total_tokens: value.total_tokens,
      avg_latency: value.count > 0 ? value.latency / value.count : 0,
    }))
    .sort((a, b) => b.count - a.count)
})
const modelStats = computed(() => aggregateBy('model'))
const providerStats = computed(() => aggregateBy('provider_name'))
const providerModelStats = computed(() => aggregateBy(['provider_name', 'actual_model_name']))

function aggregateBy(dimensions: string | string[]): any[] {
  const list = logs.value
  const SEPARATOR = '\x00'
  const groups: Record<string, {
    count: number;
    cached_tokens: number;
    input_tokens: number;
    output_tokens: number;
    total_tokens: number;
    latency: number;
    values: string[];
  }> = {}

  for (const log of list) {
    let key: string
    let values: string[]
    if (Array.isArray(dimensions)) {
      values = dimensions.map(d => String((log as any)[d] || 'unknown'))
      key = values.join(SEPARATOR)
    } else {
      values = [String((log as any)[dimensions] || 'unknown')]
      key = values[0]
    }

    if (!groups[key]) {
      groups[key] = { 
        count: 0,
        cached_tokens: 0,
        input_tokens: 0,
        output_tokens: 0,
        total_tokens: 0,
        latency: 0,
        values: values,
      }
    }
    groups[key].count++
    groups[key].cached_tokens += log.cached_tokens || 0
    groups[key].input_tokens += log.input_tokens || 0
    groups[key].output_tokens += log.output_tokens || 0
    groups[key].total_tokens += log.total_tokens || 0
    groups[key].latency += log.latency_ms || 0
  }

  return Object.entries(groups)
    .map(([key, value]) => {
      const item: any = {
        count: value.count,
        cached_tokens: value.cached_tokens,
        input_tokens: value.input_tokens,
        output_tokens: value.output_tokens,
        total_tokens: value.total_tokens,
        avg_latency: value.count > 0 ? value.latency / value.count : 0
      }
      if (Array.isArray(dimensions)) {
        dimensions.forEach((d, i) => {
          item[d] = value.values[i]
        })
      } else {
        item[dimensions] = value.values[0]
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
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res = await api.get('/usage/model-logs', { params })
    logs.value = res.data.logs || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function copyError(errorMsg: string) {
  navigator.clipboard.writeText(errorMsg).then(() => {
    ElMessage.success(t('common.copied') || 'Copied')
  }).catch(() => {
    ElMessage.error(t('common.error') || 'Error')
  })
}
</script>

<style scoped>
.usage-page { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.source-stats-card { margin-top: 20px; }
.ip-stats-card { margin-top: 20px; }
.call-method-stats-card { margin-top: 20px; }
.key-stats-card { margin-top: 20px; }
.model-stats-card { margin-top: 20px; }
.provider-stats-card { margin-top: 20px; }
.provider-model-stats-card { margin-top: 20px; }
.logs-card { margin-top: 20px; }
.error-cell { display: flex; align-items: center; }
.error-text { 
  color: var(--el-color-danger); 
  font-size: 12px;
}
</style>

<style>
.date-range-picker {
  width: 160px !important;
}
.date-range-picker .el-input__wrapper {
  width: 160px !important;
}
</style>
