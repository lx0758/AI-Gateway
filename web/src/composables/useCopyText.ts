import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

export function useCopyText() {
  const { t } = useI18n()

  function copy(text: string | undefined | null) {
    if (!text) return
    navigator.clipboard.writeText(text).then(() => {
      ElMessage.success(t('common.copied'))
    }).catch(() => {
      ElMessage.error(t('common.error'))
    })
  }

  return { copy }
}
