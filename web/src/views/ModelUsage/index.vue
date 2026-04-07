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
      <el-table :data="paginatedLogs" stripe v-loading="loading" size="small">
        <el-table-column :label="t('usage.time') || '时间'" width="180">
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
        <el-table-column prop="model" :label="t('usage.model') || '模型'" width="120">
          <template #default="{ row }">
            <span>{{ row.model }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('usage.providerModel') || '厂商/模型'" width="150">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.provider_name }}/{{ row.actual_model_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="call_method" :label="t('usage.callMethod') || '调用方式'" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.call_method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="'Tokens'" width="240">
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
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[100, 200, 500, 1000]"
        :total="logs.length"
        layout="total, sizes, prev, pager, next, jumper"
        style="margin-top: 16px; justify-content: flex-end"
      />
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
const currentPage = ref(1)
const pageSize = ref(100)

const paginatedLogs = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return logs.value.slice(start, end)
})

const stats = ref({
  totalRequests: 0,
  successRate: 0,
  totalTokens: 0,
  avgLatency: 0
})

const sourceStats = ref<any[]>([])
const ipStats = ref<any[]>([])
const callMethodStats = ref<any[]>([])
const keyStats = ref<any[]>([])
const modelStats = ref<any[]>([])
const providerStats = ref<any[]>([])
const providerModelStats = ref<any[]>([])

function getDefaultDateRange(): string[] {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0)
  const end = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1, 0, 0, 0)
  const pad = (n: number) => n.toString().padStart(2, '0')
  const format = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  return [format(start), format(end)]
}

const dateRange = ref<string[]>(getDefaultDateRange())

interface AggregationResult {
  count: number
  cached_tokens: number
  input_tokens: number
  output_tokens: number
  total_tokens: number
  latency: number
  [key: string]: any
}

function computeAllAggregations(list: LogItem[]) {
  const SEPARATOR = '\x00'
  
  if (list.length === 0) {
    stats.value = { totalRequests: 0, successRate: 0, totalTokens: 0, avgLatency: 0 }
    sourceStats.value = []
    ipStats.value = []
    callMethodStats.value = []
    keyStats.value = []
    modelStats.value = []
    providerStats.value = []
    providerModelStats.value = []
    return
  }

  const successCount = list.filter(l => l.status === 'success').length
  const totalTokens = list.reduce((sum, l) => sum + (l.total_tokens || 0), 0)
  const avgLatency = list.reduce((sum, l) => sum + (l.latency_ms || 0), 0) / list.length

  stats.value = {
    totalRequests: list.length,
    successRate: (successCount / list.length) * 100,
    totalTokens,
    avgLatency
  }

  const sourceGroups: Record<string, AggregationResult> = {}
  const ipGroups: Record<string, AggregationResult & { full_chain: string }> = {}
  const callMethodGroups: Record<string, AggregationResult> = {}
  const keyGroups: Record<string, AggregationResult> = {}
  const modelGroups: Record<string, AggregationResult> = {}
  const providerGroups: Record<string, AggregationResult> = {}
  const providerModelGroups: Record<string, AggregationResult> = {}

  for (const log of list) {
    const inc = (g: AggregationResult) => {
      g.count++
      g.cached_tokens += log.cached_tokens || 0
      g.input_tokens += log.input_tokens || 0
      g.output_tokens += log.output_tokens || 0
      g.total_tokens += log.total_tokens || 0
      g.latency += log.latency_ms || 0
    }

    const make = (): AggregationResult => ({
      count: 0, cached_tokens: 0, input_tokens: 0, output_tokens: 0, total_tokens: 0, latency: 0
    })

    if (!sourceGroups[log.source]) sourceGroups[log.source] = { ...make(), source: log.source }
    inc(sourceGroups[log.source])

    const firstIP = (log.client_ips || 'unknown').split(',')[0].trim()
    if (!ipGroups[firstIP]) ipGroups[firstIP] = { ...make(), client_ips: firstIP, full_chain: log.client_ips }
    inc(ipGroups[firstIP])

    if (!callMethodGroups[log.call_method]) callMethodGroups[log.call_method] = { ...make(), call_method: log.call_method }
    inc(callMethodGroups[log.call_method])

    if (!keyGroups[log.key_name]) keyGroups[log.key_name] = { ...make(), key_name: log.key_name }
    inc(keyGroups[log.key_name])

    if (!modelGroups[log.model]) modelGroups[log.model] = { ...make(), model: log.model }
    inc(modelGroups[log.model])

    if (!providerGroups[log.provider_name]) providerGroups[log.provider_name] = { ...make(), provider_name: log.provider_name }
    inc(providerGroups[log.provider_name])

    const pmKey = [log.provider_name, log.actual_model_name].join(SEPARATOR)
    if (!providerModelGroups[pmKey]) {
      providerModelGroups[pmKey] = { ...make(), provider_name: log.provider_name, actual_model_name: log.actual_model_name }
    }
    inc(providerModelGroups[pmKey])
  }

  const toResult = (groups: Record<string, AggregationResult>) => 
    Object.values(groups)
      .map(g => ({ ...g, avg_latency: g.count > 0 ? g.latency / g.count : 0 }))
      .sort((a, b) => b.count - a.count)

  sourceStats.value = toResult(sourceGroups)
  ipStats.value = toResult(ipGroups as any)
  callMethodStats.value = toResult(callMethodGroups)
  keyStats.value = toResult(keyGroups)
  modelStats.value = toResult(modelGroups)
  providerStats.value = toResult(providerGroups)
  providerModelStats.value = toResult(providerModelGroups)
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
    const data = res.data.logs || []
    logs.value = data
    currentPage.value = 1
    computeAllAggregations(data)
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
