<template>
  <div class="settings-page">
    <el-card>
      <template #header>{{ t('settings.changePassword') }}</template>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="150px" style="max-width: 400px">
        <el-form-item :label="t('settings.oldPassword')" prop="old_password">
          <el-input v-model="form.old_password" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('settings.newPassword')" prop="new_password">
          <el-input v-model="form.new_password" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('settings.confirmPassword')" prop="confirm_password">
          <el-input v-model="form.confirm_password" type="password" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="loading">{{ t('common.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const userStore = useUserStore()

const formRef = ref()
const loading = ref(false)

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const rules = {
  old_password: [{ required: true, message: 'Required', trigger: 'blur' }],
  new_password: [
    { required: true, message: 'Required', trigger: 'blur' },
    { min: 6, message: 'Min 6 characters', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: 'Required', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: any) => {
        if (value !== form.new_password) {
          callback(new Error('Passwords do not match'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await userStore.changePassword(form.old_password, form.new_password)
    ElMessage.success(t('settings.passwordChanged'))
    Object.assign(form, { old_password: '', new_password: '', confirm_password: '' })
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.error'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.settings-page { padding: 20px; }
</style>
