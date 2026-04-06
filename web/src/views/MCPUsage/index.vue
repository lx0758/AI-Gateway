<template>
  <div class="usage-page">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>{{ t('mcpUsage.stats') }}</span>
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            :range-separator="t('common.to')"
            :start-placeholder="t('mcpUsage.startTime')"
            :end-placeholder="t('mcpUsage.endTime')"
            value-format="YYYY-MM-DD HH:mm:ss"
            @change="fetchLogs"
            style="width: 500px; flex: 0 0 500px"
          />
        </div>
      </template>
      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic :title="t('mcpUsage.totalRequests')" :value="stats.totalRequests" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('mcpUsage.successRate')" :value="stats.successRate?.toFixed(1)" suffix="%" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('mcpUsage.totalDataSize')" :value="formatSize(stats.totalSize)" />
        </el-col>
        <el-col :span="6">
          <el-statistic :title="t('mcpUsage.avgLatency')" :value="formatLatency(stats.avgLatency)" />
        </el-col>
      </el-row>
    </el-card>

    <el-card class="source-stats-card" v-if="sourceStats.length">
      <template #header>{{ t('mcpUsage.sourceStats') }}</template>
      <el-table :data="sourceStats" stripe size="small">
        <el-table-column prop="source" :label="t('mcpUsage.source')" />
        <el-table-column prop="count" :label="t('mcpUsage.callCount')" />
        <el-table-column :label="t('mcpUsage.dataSize')">
          <template #default="{ row }">I:{{ formatSize(row.input_size) }}/O:{{ formatSize(row.output_size) }}/T:{{ formatSize(row.total_size) }}</template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.avgLatency')">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="ip-stats-card" v-if="ipStats.length">
      <template #header>{{ t('mcpUsage.ipStats') }}</template>
      <el-table :data="ipStats" stripe size="small">
        <el-table-column :label="t('mcpUsage.clientIp')">
          <template #default="{ row }">
            <span>{{ row.client_ips }}</span>
            <el-tooltip v-if="row.full_chain.includes(',')" :content="row.full_chain" placement="top">
              <el-icon style="margin-left: 4px; cursor: help;"><InfoFilled /></el-icon>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="count" :label="t('mcpUsage.callCount')" />
        <el-table-column :label="t('mcpUsage.dataSize')">
          <template #default="{ row }">I:{{ formatSize(row.input_size) }}/O:{{ formatSize(row.output_size) }}/T:{{ formatSize(row.total_size) }}</template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.avgLatency')">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="key-stats-card" v-if="keyStats.length">
      <template #header>{{ t('mcpUsage.keyStats') }}</template>
      <el-table :data="keyStats" stripe size="small">
        <el-table-column prop="key_name" :label="t('mcpUsage.keyName')" />
        <el-table-column prop="count" :label="t('mcpUsage.callCount')" />
        <el-table-column :label="t('mcpUsage.dataSize')">
          <template #default="{ row }">I:{{ formatSize(row.input_size) }}/O:{{ formatSize(row.output_size) }}/T:{{ formatSize(row.total_size) }}</template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.avgLatency')">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="mcp-stats-card" v-if="mcpStats.length">
      <template #header>{{ t('mcpUsage.mcpStats') }}</template>
      <el-table :data="mcpStats" stripe size="small">
        <el-table-column prop="mcp_name" :label="t('mcpUsage.mcpService')" />
        <el-table-column prop="mcp_type" :label="t('mcpUsage.mcpType')" />
        <el-table-column prop="count" :label="t('mcpUsage.callCount')" />
        <el-table-column :label="t('mcpUsage.dataSize')">
          <template #default="{ row }">I:{{ formatSize(row.input_size) }}/O:{{ formatSize(row.output_size) }}/T:{{ formatSize(row.total_size) }}</template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.avgLatency')">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="mcp-type-stats-card" v-if="mcpTypeStats.length">
      <template #header>{{ t('mcpUsage.mcpTypeStats') }}</template>
      <el-table :data="mcpTypeStats" stripe size="small">
        <el-table-column prop="mcp_type" :label="t('mcpUsage.mcpType')" />
        <el-table-column prop="count" :label="t('mcpUsage.callCount')" />
        <el-table-column :label="t('mcpUsage.dataSize')">
          <template #default="{ row }">I:{{ formatSize(row.input_size) }}/O:{{ formatSize(row.output_size) }}/T:{{ formatSize(row.total_size) }}</template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.avgLatency')">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="mcp-call-type-stats-card" v-if="mcpCallTypeStats.length">
      <template #header>{{ t('mcpUsage.mcpCallTypeStats') }}</template>
      <el-table :data="mcpCallTypeStats" stripe size="small">
        <el-table-column prop="call_type" :label="t('mcpUsage.callType')" />
        <el-table-column prop="count" :label="t('mcpUsage.callCount')" />
        <el-table-column :label="t('mcpUsage.dataSize')">
          <template #default="{ row }">I:{{ formatSize(row.input_size) }}/O:{{ formatSize(row.output_size) }}/T:{{ formatSize(row.total_size) }}</template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.avgLatency')">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="mcp-call-target-stats-card" v-if="mcpCallTargetStats.length">
      <template #header>{{ t('mcpUsage.mcpCallTargetStats') }}</template>
      <el-table :data="mcpCallTargetStats" stripe size="small">
        <el-table-column :label="t('mcpUsage.callType')">
          <template #default="{ row }">
            <div size="small" type="info">{{ row.call_type }}/{{ row.call_target }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="mcp_name" :label="t('mcpUsage.mcpService')" />
        <el-table-column prop="count" :label="t('mcpUsage.callCount')" />
        <el-table-column :label="t('mcpUsage.dataSize')">
          <template #default="{ row }">I:{{ formatSize(row.input_size) }}/O:{{ formatSize(row.output_size) }}/T:{{ formatSize(row.total_size) }}</template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.avgLatency')">
          <template #default="{ row }">{{ formatLatency(row.avg_latency) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="logs-card">
      <template #header>{{ t('mcpUsage.logs') }}</template>
      <el-table :data="logs" stripe v-loading="loading" size="small">
        <el-table-column :label="t('mcpUsage.time')" width="160">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="source" :label="t('mcpUsage.source')" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.source }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="client_ips" :label="t('mcpUsage.clientIp')" width="120">
          <template #default="{ row }">
            <span>{{ row.client_ips.split(',')[0].trim() }}</span>
            <el-tooltip v-if="row.client_ips.includes(',')" :content="row.client_ips" placement="top">
              <el-icon style="margin-left: 4px; cursor: help;"><InfoFilled /></el-icon>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="key_name" :label="t('mcpUsage.keyName')" width="100" />
        <el-table-column prop="mcp_name" :label="t('mcpUsage.mcpService')" width="200">
          <template #default="{ row }">
            <span>{{ row.mcp_name }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.mcpType')" width="80">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.mcp_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.callTarget')" width="240">
          <template #default="{ row }">
            <span>{{ row.call_type }}/{{ row.call_target }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('mcpUsage.dataSize')" width="180">
          <template #default="{ row }">I:{{ formatSize(row.input_size) }}/O:{{ formatSize(row.output_size) }}</template>
        </el-table-column>
        <el-table-column prop="latency_ms" :label="t('mcpUsage.latency')" width="80">
          <template #default="{ row }">{{ formatLatency(row.latency_ms) }}</template>
        </el-table-column>
        <el-table-column prop="status" :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="error_msg" :label="t('mcpUsage.errorMsg')" show-overflow-tooltip>
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
import { formatDateTime, formatLatency } from '@/utils/format'

const { t } = useI18n()

interface LogItem {
  id: number
  source: string
  client_ips: string
  key_id: number
  key_name: string
  mcp_id: number
  mcp_name: string
  mcp_type: string
  call_type: string
  call_method: string
  call_target: string
  input_size: number
  output_size: number
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

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const stats = computed(() => {
  const list = logs.value
  if (list.length === 0) {
    return {
      totalRequests: 0,
      successRate: 0,
      totalSize: 0,
      avgLatency: 0,
    }
  }
  const totalRequests = list.length
  const successCount = list.filter(l => l.status === 'success').length
  const successRate = (successCount / totalRequests) * 100
  const totalSize = list.reduce((sum, l) => sum + (l.input_size || 0) + (l.output_size || 0), 0)
  const avgLatency = list.reduce((sum, l) => sum + (l.latency_ms || 0), 0) / list.length
  return {
    totalRequests,
    successRate,
    totalSize,
    avgLatency
  }
})

const sourceStats = computed(() => aggregateBy('source'))
const keyStats = computed(() => aggregateBy('key_name'))
const mcpStats = computed(() => aggregateBy(['mcp_name', 'mcp_type']))
const mcpTypeStats = computed(() => aggregateBy('mcp_type'))
const mcpCallTypeStats = computed(() => aggregateBy('call_type'))
const mcpCallTargetStats = computed(() => aggregateBy(['call_type', 'call_target', 'mcp_name']))

const ipStats = computed(() => {
  const list = logs.value
  const groups: Record<string, {
    count: number;
    input_size: number;
    output_size: number;
    total_size: number;
    latency: number;
    full_chain: string;
  }> = {}

  for (const log of list) {
    const chain = log.client_ips || 'unknown'
    const firstIP = chain.split(',')[0].trim()
    if (!groups[firstIP]) {
      groups[firstIP] = {
        count: 0,
        input_size: 0,
        output_size: 0,
        total_size: 0,
        latency: 0,
        full_chain: chain,
      }
    }
    groups[firstIP].count++
    groups[firstIP].input_size += log.input_size || 0
    groups[firstIP].output_size += log.output_size || 0
    groups[firstIP].total_size += (log.input_size || 0) + (log.output_size || 0)
    groups[firstIP].latency += log.latency_ms || 0
  }

  return Object.entries(groups)
    .map(([key, value]) => ({
      client_ips: key,
      full_chain: value.full_chain,
      count: value.count,
      input_size: value.input_size,
      output_size: value.output_size,
      total_size: value.total_size,
      avg_latency: value.count > 0 ? value.latency / value.count : 0,
    }))
    .sort((a, b) => b.count - a.count)
})

function aggregateBy(dimensions: string | string[]): any[] {
  const list = logs.value
  const SEPARATOR = '\x00'
  const groups: Record<string, {
    count: number;
    input_size: number;
    output_size: number;
    total_size: number;
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
        input_size: 0,
        output_size: 0,
        total_size: 0,
        latency: 0,
        values: values,
      }
    }
    groups[key].count++
    groups[key].input_size += log.input_size || 0
    groups[key].output_size += log.output_size || 0
    groups[key].total_size += (log.input_size || 0) + (log.output_size || 0)
    groups[key].latency += log.latency_ms || 0
  }

  return Object.entries(groups)
    .map(([key, value]) => {
      const item: any = {
        count: value.count,
        input_size: value.input_size,
        output_size: value.output_size,
        total_size: value.total_size,
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
    const res = await api.get('/usage/mcp-logs', { params })
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
.key-stats-card { margin-top: 20px; }
.mcp-stats-card { margin-top: 20px; }
.mcp-type-stats-card { margin-top: 20px; }
.mcp-call-type-stats-card { margin-top: 20px; }
.mcp-call-target-stats-card { margin-top: 20px; }
.logs-card { margin-top: 20px; }
.error-cell { display: flex; align-items: center; }
.error-text { 
  color: var(--el-color-danger); 
  font-size: 12px;
}
</style>