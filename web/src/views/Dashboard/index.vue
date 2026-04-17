<template>
  <div class="dashboard">
    <el-card class="section-card">
      <template #header>
        <div class="section-title">
          <el-icon><Box /></el-icon>
          <span>{{ t('dashboard.assetsOverview') || '资产概览' }}</span>
        </div>
      </template>
      <el-row :gutter="20">
        <el-col :xs="12" :sm="6" :md="6" class="stat-col">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value primary">{{ stats.assets.activeProviders }}/{{ stats.assets.totalProviders }}</div>
            <div class="stat-label">{{ t('dashboard.providers') || '厂商' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6" class="stat-col">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value primary">{{ stats.assets.activeProviderModels }}/{{ stats.assets.totalProviderModels }}</div>
            <div class="stat-label">{{ t('dashboard.providerModels') || '厂商模型' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6" class="stat-col">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value primary">{{ stats.assets.activeModels }}/{{ stats.assets.totalModels }}</div>
            <div class="stat-label">{{ t('dashboard.models') || '模型' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6" class="stat-col">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value success">{{ stats.assets.activeKeys }}/{{ stats.assets.totalKeys }}</div>
            <div class="stat-label">{{ t('dashboard.keys') || 'Keys' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6" class="stat-col">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value warning">{{ stats.assets.activeMCPs }}/{{ stats.assets.totalMCPs }}</div>
            <div class="stat-label">{{ t('dashboard.mcpServices') || 'MCP服务' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6" class="stat-col">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value warning">{{ stats.assets.activeMCPTools }}/{{ stats.assets.totalMCPTools }}</div>
            <div class="stat-label">{{ t('dashboard.mcpTools') || 'MCP工具' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6" class="stat-col">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value warning">{{ stats.assets.activeMCPResources }}/{{ stats.assets.totalMCPResources }}</div>
            <div class="stat-label">{{ t('dashboard.mcpResources') || 'MCP资源' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6" class="stat-col">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value warning">{{ stats.assets.activeMCPPrompts }}/{{ stats.assets.totalMCPPrompts }}</div>
            <div class="stat-label">{{ t('dashboard.mcpPrompts') || 'MCP提示词' }}</div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card class="section-card">
      <template #header>
        <div class="section-title">
          <el-icon><Cpu /></el-icon>
          <span>{{ t('dashboard.modelUsage') || 'Model API 使用情况' }}</span>
          <el-tag size="small" type="info">{{ t('dashboard.lastNDays', { n: stats.days }) || `过去${stats.days}天` }}</el-tag>
        </div>
      </template>
      <el-row :gutter="20" class="stats-row">
        <el-col :xs="12" :sm="6" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value">{{ formatNumber(stats.modelUsage.totalRequests) }}</div>
            <div class="stat-label">{{ t('dashboard.totalRequests') || '总请求' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value primary">{{ formatTokens(stats.modelUsage.totalTokens) }}</div>
            <div class="stat-label">{{ t('dashboard.totalTokens') || '总Tokens' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value success">{{ stats.modelUsage.successCount > 0 ? ((stats.modelUsage.successCount / stats.modelUsage.totalRequests) * 100).toFixed(1) : 0 }}%</div>
            <div class="stat-label">{{ t('dashboard.successRate') || '成功率' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value warning">{{ formatLatency(stats.modelUsage.avgLatency) }}</div>
            <div class="stat-label">{{ t('dashboard.avgLatency') || '平均耗时' }}</div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="charts-row">
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover">
            <template #header>{{ t('dashboard.requestTrend') || '请求趋势' }}</template>
            <div v-if="hasModelTrendData" ref="modelTrendChartRef" class="chart" v-loading="loading"></div>
            <el-empty v-else :description="t('common.noData')" />
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover">
            <template #header>{{ t('dashboard.tokenTrend') || 'Token消耗趋势' }}</template>
            <div v-if="hasTokenTrendData" ref="tokenTrendChartRef" class="chart" v-loading="loading"></div>
            <el-empty v-else :description="t('common.noData')" />
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="charts-row">
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover">
            <template #header>{{ t('dashboard.providerDistribution') || '厂商分布' }}</template>
            <div v-if="hasModelProviderData" ref="modelPieChartRef" class="chart" v-loading="loading"></div>
            <el-empty v-else :description="t('common.noData')" />
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover">
            <template #header>{{ t('dashboard.modelDistribution') || '模型分布' }}</template>
            <div v-if="hasModelStatsData" ref="modelBarChartRef" class="chart" v-loading="loading"></div>
            <el-empty v-else :description="t('common.noData')" />
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card class="section-card">
      <template #header>
        <div class="section-title">
          <el-icon><Connection /></el-icon>
          <span>{{ t('dashboard.mcpUsage') || 'MCP 服务使用情况' }}</span>
          <el-tag size="small" type="info">{{ t('dashboard.lastNDays', { n: stats.days }) || `过去${stats.days}天` }}</el-tag>
        </div>
      </template>
      <el-row :gutter="20" class="stats-row">
        <el-col :xs="12" :sm="6" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value">{{ formatNumber(stats.mcpUsage.totalRequests) }}</div>
            <div class="stat-label">{{ t('dashboard.totalCalls') || '总调用' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value primary">{{ formatSize(stats.mcpUsage.totalSize) }}</div>
            <div class="stat-label">{{ t('dashboard.totalDataSize') || '总数据量' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value success">{{ stats.mcpUsage.successCount > 0 ? ((stats.mcpUsage.successCount / stats.mcpUsage.totalRequests) * 100).toFixed(1) : 0 }}%</div>
            <div class="stat-label">{{ t('dashboard.successRate') || '成功率' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :md="6">
          <el-card shadow="hover" class="stat-card">
            <div class="stat-value warning">{{ formatLatency(stats.mcpUsage.avgLatency) }}</div>
            <div class="stat-label">{{ t('dashboard.avgLatency') || '平均耗时' }}</div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="charts-row">
        <el-col :xs="24" :sm="16">
          <el-card shadow="hover">
            <template #header>{{ t('dashboard.callTrend') || '调用趋势' }}</template>
            <div v-if="hasMCPTrendData" ref="mcpTrendChartRef" class="chart" v-loading="loading"></div>
            <el-empty v-else :description="t('common.noData')" />
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="8">
          <el-card shadow="hover">
            <template #header>{{ t('dashboard.mcpTypeDistribution') || '类型分布' }}</template>
            <div v-if="hasMCPTypeData" ref="mcpPieChartRef" class="chart" v-loading="loading"></div>
            <el-empty v-else :description="t('common.noData')" />
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import * as echarts from 'echarts'
import { Box, Cpu, Connection } from '@element-plus/icons-vue'
import api from '@/api'
import { formatLatency, formatTokens } from '@/utils/format'

const { t } = useI18n()
const isDark = computed(() => document.documentElement.classList.contains('dark'))

interface AssetStats {
  totalProviders: number
  activeProviders: number
  totalModels: number
  activeModels: number
  totalProviderModels: number
  activeProviderModels: number
  totalMCPs: number
  activeMCPs: number
  totalKeys: number
  activeKeys: number
  totalMCPTools: number
  activeMCPTools: number
  totalMCPResources: number
  activeMCPResources: number
  totalMCPPrompts: number
  activeMCPPrompts: number
}

interface ModelUsage {
  totalRequests: number
  successCount: number
  totalTokens: number
  avgLatency: number
  dailyStats: Array<{ date: string; count: number; success: number }>
  tokenDailyStats: Array<{ date: string; total_tokens: number }>
  providerStats: Array<{ provider: string; count: number; tokens: number; avg_latency: number }>
  modelStats: Array<{ model: string; count: number }>
  providerModelStats: Array<{ provider: string; model_count: number }>
}

interface MCPUsage {
  totalRequests: number
  successCount: number
  totalSize: number
  avgLatency: number
  dailyStats: Array<{ date: string; count: number; success: number }>
  typeStats: Array<{ mcp_type: string; count: number }>
  serviceStats: Array<{ mcp_name: string; count: number }>
}

const stats = ref<{
  days: number
  assets: AssetStats
  modelUsage: ModelUsage
  mcpUsage: MCPUsage
}>({
  days: 7,
  assets: {
    totalProviders: 0,
    activeProviders: 0,
    totalModels: 0,
    activeModels: 0,
    totalProviderModels: 0,
    activeProviderModels: 0,
    totalMCPs: 0,
    activeMCPs: 0,
    totalKeys: 0,
    activeKeys: 0,
    totalMCPTools: 0,
    activeMCPTools: 0,
    totalMCPResources: 0,
    activeMCPResources: 0,
    totalMCPPrompts: 0,
    activeMCPPrompts: 0
  },
  modelUsage: {
    totalRequests: 0,
    successCount: 0,
    totalTokens: 0,
    avgLatency: 0,
    dailyStats: [],
    tokenDailyStats: [],
    providerStats: [],
    providerModelStats: [],
    modelStats: []
  },
  mcpUsage: {
    totalRequests: 0,
    successCount: 0,
    totalSize: 0,
    avgLatency: 0,
    dailyStats: [],
    typeStats: [],
    serviceStats: []
  }
})

const loading = ref(false)
const modelTrendChartRef = ref<HTMLElement>()
const tokenTrendChartRef = ref<HTMLElement>()
const modelPieChartRef = ref<HTMLElement>()
const modelBarChartRef = ref<HTMLElement>()
const mcpTrendChartRef = ref<HTMLElement>()
const mcpPieChartRef = ref<HTMLElement>()

let modelTrendChart: echarts.ECharts | null = null
let tokenTrendChart: echarts.ECharts | null = null
let modelPieChart: echarts.ECharts | null = null
let modelBarChart: echarts.ECharts | null = null
let mcpTrendChart: echarts.ECharts | null = null
let mcpPieChart: echarts.ECharts | null = null

const hasModelTrendData = computed(() => stats.value.modelUsage.dailyStats && stats.value.modelUsage.dailyStats.length > 0)
const hasTokenTrendData = computed(() => stats.value.modelUsage.tokenDailyStats && stats.value.modelUsage.tokenDailyStats.length > 0)
const hasModelProviderData = computed(() => stats.value.modelUsage.providerStats && stats.value.modelUsage.providerStats.length > 0)
const hasModelStatsData = computed(() => stats.value.modelUsage.modelStats && stats.value.modelUsage.modelStats.length > 0)
const hasMCPTrendData = computed(() => stats.value.mcpUsage.dailyStats && stats.value.mcpUsage.dailyStats.length > 0)
const hasMCPTypeData = computed(() => stats.value.mcpUsage.typeStats && stats.value.mcpUsage.typeStats.length > 0)

onMounted(async () => {
  await fetchDashboard()
  initCharts()
})

watch(isDark, () => {
  initCharts()
})

watch([hasModelTrendData, hasTokenTrendData, hasModelProviderData, hasModelStatsData, hasMCPTrendData, hasMCPTypeData], async () => {
  await nextTick()
  initCharts()
})

async function fetchDashboard() {
  loading.value = true
  try {
    const res = await api.get('/usage/dashboard')
    stats.value = res.data
    await nextTick()
    initCharts()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function initCharts() {
  if (hasModelTrendData.value) initModelTrendChart()
  if (hasTokenTrendData.value) initTokenTrendChart()
  if (hasModelProviderData.value) initModelPieChart()
  if (hasModelStatsData.value) initModelBarChart()
  if (hasMCPTrendData.value) initMCPTrendChart()
  if (hasMCPTypeData.value) initMCPPieChart()
}

function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatChartDate(date: string): string {
  if (!date || date.length < 10) return date
  const [year, month, day] = date.split('-')
  return `${month}/${day}`
}

function getChartTheme() {
  return {
    textColor: isDark.value ? '#a3a3a3' : '#666',
    bgColor: isDark.value ? '#1f1f1f' : '#fff',
    borderColor: isDark.value ? '#333' : '#ddd',
    textColorDark: isDark.value ? '#fff' : '#333',
    splitLine: isDark.value ? '#333' : '#eee'
  }
}

function initModelTrendChart() {
  if (!modelTrendChartRef.value) return
  
  if (!modelTrendChart) {
    modelTrendChart = echarts.init(modelTrendChartRef.value)
  }
  
  const theme = getChartTheme()
  const data = stats.value.modelUsage.dailyStats
  
  modelTrendChart.setOption({
    tooltip: { 
      trigger: 'axis',
      backgroundColor: theme.bgColor,
      borderColor: theme.borderColor,
      textStyle: { color: theme.textColorDark }
    },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: data.map((d: any) => formatChartDate(d.date)),
      axisLine: { lineStyle: { color: theme.textColor } },
      axisLabel: { color: theme.textColor }
    },
    yAxis: { 
      type: 'value',
      axisLine: { lineStyle: { color: theme.textColor } },
      axisLabel: { color: theme.textColor },
      splitLine: { lineStyle: { color: theme.splitLine } }
    },
    series: [{
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      lineStyle: { width: 2, color: '#409eff' },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
          { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
        ])
      },
      itemStyle: { color: '#409eff' },
      data: data.map((d: any) => d.count)
    }]
  })
}

function initTokenTrendChart() {
  if (!tokenTrendChartRef.value) return
  
  if (!tokenTrendChart) {
    tokenTrendChart = echarts.init(tokenTrendChartRef.value)
  }
  
  const theme = getChartTheme()
  const data = stats.value.modelUsage.tokenDailyStats
  
  tokenTrendChart.setOption({
    tooltip: { 
      trigger: 'axis',
      backgroundColor: theme.bgColor,
      borderColor: theme.borderColor,
      textStyle: { color: theme.textColorDark }
    },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: data.map((d: any) => formatChartDate(d.date)),
      axisLine: { lineStyle: { color: theme.textColor } },
      axisLabel: { color: theme.textColor }
    },
    yAxis: { 
      type: 'value',
      axisLine: { lineStyle: { color: theme.textColor } },
      axisLabel: { 
        color: theme.textColor,
        formatter: (value: number) => {
          if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M'
          if (value >= 1000) return (value / 1000).toFixed(1) + 'K'
          return value.toString()
        }
      },
      splitLine: { lineStyle: { color: theme.splitLine } }
    },
    series: [{
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      lineStyle: { width: 2, color: '#67c23a' },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(103, 194, 58, 0.3)' },
          { offset: 1, color: 'rgba(103, 194, 58, 0.05)' }
        ])
      },
      itemStyle: { color: '#67c23a' },
      data: data.map((d: any) => d.total_tokens)
    }]
  })
}

function initModelPieChart() {
  if (!modelPieChartRef.value) return
  
  if (!modelPieChart) {
    modelPieChart = echarts.init(modelPieChartRef.value)
  }
  
  const theme = getChartTheme()
  const data = stats.value.modelUsage.providerStats
  
  modelPieChart.setOption({
    tooltip: { 
      trigger: 'item',
      backgroundColor: theme.bgColor,
      borderColor: theme.borderColor,
      textStyle: { color: theme.textColorDark }
    },
    legend: {
      orient: 'vertical',
      right: '5%',
      top: 'center',
      textStyle: { color: theme.textColor }
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
      labelLine: { show: false },
      data: data.map((p: any) => ({ name: p.provider, value: p.count }))
    }]
  })
}

function initModelBarChart() {
  if (!modelBarChartRef.value) return
  
  if (!modelBarChart) {
    modelBarChart = echarts.init(modelBarChartRef.value)
  }
  
  const theme = getChartTheme()
  const data = stats.value.modelUsage.modelStats.slice(0, 10)
  
  modelBarChart.setOption({
    tooltip: { 
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: theme.bgColor,
      borderColor: theme.borderColor,
      textStyle: { color: theme.textColorDark }
    },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { 
      type: 'category',
      data: data.map((d: any) => d.model),
      axisLine: { lineStyle: { color: theme.textColor } },
      axisLabel: { color: theme.textColor, rotate: 30 }
    },
    yAxis: { 
      type: 'value',
      axisLine: { lineStyle: { color: theme.textColor } },
      axisLabel: { color: theme.textColor },
      splitLine: { lineStyle: { color: theme.splitLine } }
    },
    series: [{
      type: 'bar',
      data: data.map((d: any) => d.count),
      itemStyle: { color: '#409eff' }
    }]
  })
}

function initMCPTrendChart() {
  if (!mcpTrendChartRef.value) return
  
  if (!mcpTrendChart) {
    mcpTrendChart = echarts.init(mcpTrendChartRef.value)
  }
  
  const theme = getChartTheme()
  const data = stats.value.mcpUsage.dailyStats
  
  mcpTrendChart.setOption({
    tooltip: { 
      trigger: 'axis',
      backgroundColor: theme.bgColor,
      borderColor: theme.borderColor,
      textStyle: { color: theme.textColorDark }
    },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: data.map((d: any) => formatChartDate(d.date)),
      axisLine: { lineStyle: { color: theme.textColor } },
      axisLabel: { color: theme.textColor }
    },
    yAxis: { 
      type: 'value',
      axisLine: { lineStyle: { color: theme.textColor } },
      axisLabel: { color: theme.textColor },
      splitLine: { lineStyle: { color: theme.splitLine } }
    },
    series: [{
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      lineStyle: { width: 2, color: '#e6a23c' },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(230, 162, 60, 0.3)' },
          { offset: 1, color: 'rgba(230, 162, 60, 0.05)' }
        ])
      },
      itemStyle: { color: '#e6a23c' },
      data: data.map((d: any) => d.count)
    }]
  })
}

function initMCPPieChart() {
  if (!mcpPieChartRef.value) return
  
  if (!mcpPieChart) {
    mcpPieChart = echarts.init(mcpPieChartRef.value)
  }
  
  const theme = getChartTheme()
  const data = stats.value.mcpUsage.typeStats
  
  mcpPieChart.setOption({
    tooltip: { 
      trigger: 'item',
      backgroundColor: theme.bgColor,
      borderColor: theme.borderColor,
      textStyle: { color: theme.textColorDark }
    },
    legend: {
      orient: 'vertical',
      right: '5%',
      top: 'center',
      textStyle: { color: theme.textColor }
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
      labelLine: { show: false },
      data: data.map((p: any) => ({ name: p.mcp_type, value: p.count }))
    }]
  })
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.section-card {
  margin-bottom: 20px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.stat-card {
  text-align: center;
  transition: box-shadow 0.3s ease;
}

.stat-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--el-color-primary);
}

.stat-value.success {
  color: var(--el-color-success);
}

.stat-value.primary {
  color: var(--el-color-primary);
}

.stat-value.warning {
  color: var(--el-color-warning);
}

.stat-value.danger {
  color: var(--el-color-danger);
}

.stat-value.info {
  color: var(--el-color-info);
}

.stat-label {
  margin-top: 8px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.stats-row {
  margin-bottom: 15px;
}

.charts-row {
  margin-bottom: 15px;
}

.stat-col {
  margin-bottom: 15px;
}

.chart {
  height: 280px;
}
</style>
