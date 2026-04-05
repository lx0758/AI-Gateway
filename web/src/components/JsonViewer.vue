<template>
  <div class="json-viewer">
    <div class="json-header">
      <span class="json-title">{{ title }}</span>
      <el-button link type="primary" @click="copyJson" size="small">
        <el-icon><CopyDocument /></el-icon>
        {{ t('mcp.copyJson') }}
      </el-button>
    </div>
    <div class="json-content">
      <VueJsonPretty
        :data="parsedJson"
        :deep="3"
        :show-line="true"
        :show-double-quotes="true"
        :show-icon="true"
        :collapsed-on-click-brackets="true"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'

const props = defineProps<{
  json: string
  title?: string
}>()

const { t } = useI18n()

const parsedJson = computed(() => {
  if (!props.json) return null
  try {
    return JSON.parse(props.json)
  } catch {
    return props.json
  }
})

const formattedJson = computed(() => {
  if (!props.json) return ''
  try {
    return JSON.stringify(JSON.parse(props.json), null, 2)
  } catch {
    return props.json
  }
})

function copyJson() {
  if (!formattedJson.value) return
  navigator.clipboard.writeText(formattedJson.value).then(() => {
    ElMessage.success(t('common.copied'))
  }).catch(() => {
    ElMessage.error(t('common.error'))
  })
}
</script>

<style scoped>
.json-viewer {
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  overflow: hidden;
}

.json-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background-color: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color);
}

.json-title {
  font-weight: 500;
  font-size: 14px;
}

.json-content {
  padding: 12px;
  background-color: var(--el-bg-color);
  max-height: 500px;
  overflow: auto;
}

.json-content :deep(.vjs-tree) {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.json-content :deep(.vjs-key) {
  color: #c678dd;
}

.json-content :deep(.vjs-value-string) {
  color: #98c379;
}

.json-content :deep(.vjs-value-number) {
  color: #d19a66;
}

.json-content :deep(.vjs-value-boolean) {
  color: #56b6c2;
}

.json-content :deep(.vjs-value-null) {
  color: #e06c75;
}
</style>
