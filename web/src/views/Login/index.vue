<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="title">{{ t('login.title') }}</h1>
      <el-form :model="form" :rules="rules" ref="formRef" @submit.prevent="handleLogin">
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            :placeholder="t('login.username')"
            size="large"
            prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            :placeholder="t('login.password')"
            size="large"
            prefix-icon="Lock"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            native-type="submit"
            :loading="loading"
            class="login-btn"
          >
            {{ t('login.submit') }}
          </el-button>
        </el-form-item>
      </el-form>
      <div class="lang-switch">
        <el-button text @click="changeLocale('zh')" :class="{ active: locale === 'zh' }">中文</el-button>
        <el-button text @click="changeLocale('en')" :class="{ active: locale === 'en' }">EN</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'

const { t, locale } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()

const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: () => t('login.username'), trigger: 'blur' }],
  password: [{ required: true, message: () => t('login.password'), trigger: 'blur' }]
}

async function handleLogin() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const success = await userStore.login(form.username, form.password)
    if (success) {
      router.push('/')
    } else {
      ElMessage.error(t('login.invalidCredentials'))
    }
  } finally {
    loading.value = false
  }
}

function changeLocale(lang: string) {
  locale.value = lang
  appStore.setLocale(lang)
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  padding: 40px;
  background: var(--el-bg-color);
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.title {
  text-align: center;
  margin-bottom: 30px;
  font-size: 24px;
  color: var(--el-text-color-primary);
}

.login-btn {
  width: 100%;
}

.lang-switch {
  display: flex;
  justify-content: center;
  gap: 10px;
  margin-top: 20px;
}

.lang-switch .active {
  color: var(--el-color-primary);
  font-weight: bold;
}
</style>
