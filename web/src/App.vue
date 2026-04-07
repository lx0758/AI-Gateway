<script setup lang="ts">
import { computed } from 'vue'
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import { useAppStore } from './stores/app'

const appStore = useAppStore()
const elementLocale = computed(() => appStore.locale === 'en' ? en : zhCn)
</script>

<template>
  <el-config-provider :locale="elementLocale">
    <router-view />
  </el-config-provider>
</template>
