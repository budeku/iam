<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const userStore = useUserStore()
const { me, loading } = storeToRefs(userStore)

onMounted(() => {
  if (!me.value) {
    userStore.getMe()
  }
})
</script>

<template>
  <div class="dashboard-container" v-loading="loading">
    <h1 v-if="me" class="greeting-title">
      {{ t('dashboard.greeting', { name: me.username }) }}
    </h1>

    <h1 v-else-if="!loading" class="greeting-title">Dashboard</h1>
  </div>
</template>

<style scoped lang="scss">
.dashboard-container {
  padding: 24px;

  .greeting-title {
    font-size: 28px;
    font-weight: 600;
    color: var(--text-color);
    margin-bottom: 20px;
  }
}
</style>
