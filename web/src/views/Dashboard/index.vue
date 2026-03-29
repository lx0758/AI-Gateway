<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.totalRequests || 0 }}</div>
          <div class="stat-label">{{ t('dashboard.totalRequests') }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.todayRequests || 0 }}</div>
          <div class="stat-label">{{ t('dashboard.todayRequests') }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.activeProviders || 0 }}</div>
          <div class="stat-label">{{ t('dashboard.activeProviders') }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.activeKeys || 0 }}</div>
          <div class="stat-label">{{ t('dashboard.activeKeys') }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="charts-row">
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>{{ t('dashboard.requestTrend') }}</template>
          <div v-if="hasTrendData" ref="trendChartRef" class="chart" v-loading="loading"></div>
          <el-empty v-else :description="t('common.noData')" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>{{ t('dashboard.providerDistribution') }}</template>
          <div v-if="hasProviderData" ref="pieChartRef" class="chart" v-loading="loading"></div>
          <el-empty v-else :description="t('common.noData')" />
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover">
      <template #header>{{ t('dashboard.modelRanking') }}</template>
      <el-table v-if="hasModelData" :data="stats.modelStats || []" stripe v-loading="loading">
        <el-table-column prop="model" :label="t('common.name')" />
        <el-table-column prop="count" :label="t('common.total')">
          <template #default="{ row }">
            {{ row.count }} {{ t('common.type') === 'Type' ? 'requests' : '次' }}
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else :description="t('common.noData')" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import * as echarts from 'echarts'
import api from '@/api'

const { t } = useI18n()
const isDark = computed(() => document.documentElement.classList.contains('dark'))

const stats = ref<any>({
  totalRequests: 0,
  todayRequests: 0,
  activeProviders: 0,
  activeKeys: 0,
  dailyStats: [],
  providerStats: [],
  modelStats: []
})

const loading = ref(false)
const trendChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()
let trendChart: echarts.ECharts | null = null
let pieChart: echarts.ECharts | null = null

const hasTrendData = computed(() => stats.value.dailyStats && stats.value.dailyStats.length > 0)
const hasProviderData = computed(() => stats.value.providerStats && stats.value.providerStats.length > 0)
const hasModelData = computed(() => stats.value.modelStats && stats.value.modelStats.length > 0)

onMounted(async () => {
  await fetchDashboard()
  initCharts()
})

watch(isDark, () => {
  initCharts()
})

watch([hasTrendData, hasProviderData], async () => {
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
  if (hasTrendData.value) initTrendChart()
  if (hasProviderData.value) initPieChart()
}

function initTrendChart() {
  if (!trendChartRef.value) return
  
  if (!trendChart) {
    trendChart = echarts.init(trendChartRef.value)
  }
  
  const hasData = stats.value.dailyStats && stats.value.dailyStats.length > 0
  const textColor = isDark.value ? '#a3a3a3' : '#666'
  const lineColor = '#409eff'
  
  trendChart.setOption({
    tooltip: { 
      trigger: 'axis',
      backgroundColor: isDark.value ? '#1f1f1f' : '#fff',
      borderColor: isDark.value ? '#333' : '#ddd',
      textStyle: { color: isDark.value ? '#fff' : '#333' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: hasData ? stats.value.dailyStats.map((d: any) => d.date) : [],
      axisLine: { lineStyle: { color: textColor } },
      axisLabel: { color: textColor }
    },
    yAxis: { 
      type: 'value',
      axisLine: { lineStyle: { color: textColor } },
      axisLabel: { color: textColor },
      splitLine: { lineStyle: { color: isDark.value ? '#333' : '#eee' } }
    },
    series: [{
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 8,
      lineStyle: { width: 3, color: lineColor },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
          { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
        ])
      },
      itemStyle: { color: lineColor },
      data: hasData ? stats.value.dailyStats.map((d: any) => d.count) : []
    }]
  })
}

function initPieChart() {
  if (!pieChartRef.value) return
  
  if (!pieChart) {
    pieChart = echarts.init(pieChartRef.value)
  }
  
  const hasData = stats.value.providerStats && stats.value.providerStats.length > 0
  const textColor = isDark.value ? '#a3a3a3' : '#666'
  
  pieChart.setOption({
    tooltip: { 
      trigger: 'item',
      backgroundColor: isDark.value ? '#1f1f1f' : '#fff',
      borderColor: isDark.value ? '#333' : '#ddd',
      textStyle: { color: isDark.value ? '#fff' : '#333' }
    },
    legend: {
      orient: 'vertical',
      right: '5%',
      top: 'center',
      textStyle: { color: textColor }
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      label: { show: false },
      emphasis: {
        label: { show: true, fontSize: 14, fontWeight: 'bold' }
      },
      labelLine: { show: false },
      data: hasData ? stats.value.providerStats.map((p: any) => ({
        name: p.provider,
        value: p.count
      })) : []
    }]
  })
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.stat-card {
  text-align: center;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: var(--el-color-primary);
}

.stat-label {
  margin-top: 8px;
  color: var(--el-text-color-secondary);
}

.charts-row {
  margin-top: 20px;
  margin-bottom: 20px;
}

.chart {
  height: 300px;
}
</style>
