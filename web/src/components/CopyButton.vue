<template>
  <el-button v-if="text" link type="primary" size="small" @click.stop="handleCopy">
    <el-icon><CopyDocument /></el-icon>
  </el-button>
</template>

<script setup lang="ts">
import { CopyDocument } from '@element-plus/icons-vue'
import { useCopyText } from '@/composables/useCopyText'

const props = defineProps<{
  text?: string | null
}>()

const { copy } = useCopyText()

function handleCopy() {
  copy(props.text)
}
</script>
