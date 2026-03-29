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
          <div ref="trendChartRef" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>{{ t('dashboard.providerDistribution') }}</template>
          <div ref="pieChartRef" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover">
      <template #header>{{ t('dashboard.modelRanking') }}</template>
      <el-table :data="stats.modelStats || []" stripe>
        <el-table-column prop="model" :label="t('common.name')" />
        <el-table-column prop="count" :label="t('common.total')">
          <template #default="{ row }">
            {{ row.count }} {{ t('common.type') === 'Type' ? 'requests' : '次' }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import * as echarts from 'echarts'
import api from '@/api'

const { t } = useI18n()

const stats = ref<any>({
  totalRequests: 0,
  todayRequests: 0,
  activeProviders: 0,
  activeKeys: 0,
  dailyStats: [],
  providerStats: [],
  modelStats: []
})

const trendChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()

onMounted(async () => {
  await fetchDashboard()
  initCharts()
})

async function fetchDashboard() {
  try {
    const res = await api.get('/usage/dashboard')
    stats.value = res.data
  } catch (e) {
    console.error(e)
  }
}

function initCharts() {
  if (trendChartRef.value) {
    const chart = echarts.init(trendChartRef.value)
    chart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: {
        type: 'category',
        data: stats.value.dailyStats.map((d: any) => d.date)
      },
      yAxis: { type: 'value' },
      series: [{
        type: 'line',
        smooth: true,
        data: stats.value.dailyStats.map((d: any) => d.count)
      }]
    })
  }

  if (pieChartRef.value) {
    const chart = echarts.init(pieChartRef.value)
    chart.setOption({
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie',
        radius: '60%',
        data: stats.value.providerStats.map((p: any) => ({
          name: p.provider,
          value: p.count
        }))
      }]
    })
  }
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
