<script setup lang="ts">
import { onMounted } from 'vue'
import { useProfile } from './components/useProfile'
import type { UserSessionItemResp } from '@/api/modules/user'
import { useI18n } from 'vue-i18n'
import { formatDate } from '@/utils/date'

const { t } = useI18n()
const { loading, userInfo, sessionList, revokeSession, init } = useProfile()

function toRow(row: unknown): UserSessionItemResp {
  return row as UserSessionItemResp
}

onMounted(() => {
  init()
})
</script>

<template>
  <div class="profile-container" v-loading="loading">
    <el-card class="box-card mb-4" shadow="hover">
      <template #header>
        <div class="card-header">
          <span>{{ t('user.profile.basicInfo') }}</span>
        </div>
      </template>
      <el-descriptions :column="2" border v-if="userInfo">
        <el-descriptions-item :label="t('user.profile.username')">
          <el-tag size="small">{{ userInfo.username }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('user.profile.email')">
          {{ userInfo.email }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('user.profile.uuid')">
          <span class="text-gray-400">{{ userInfo.uuid }}</span>
        </el-descriptions-item>
        <el-descriptions-item :label="t('user.profile.accountType')">
          <el-tag :type="userInfo.is_super_admin ? 'danger' : 'success'" size="small">
            {{
              userInfo.is_super_admin ? t('user.profile.superAdmin') : t('user.profile.regularUser')
            }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card class="box-card" shadow="hover">
      <template #header>
        <div class="card-header flex justify-between items-center">
          <span>{{ t('user.profile.sessionManagement') }}</span>
          <el-button size="default" @click="init">{{ t('user.profile.refresh') }}</el-button>
        </div>
      </template>

      <el-table
        :data="sessionList"
        style="width: 100%"
        border
        stripe
        :header-cell-style="{ textAlign: 'center' }"
        :cell-style="{ textAlign: 'center' }"
      >
        <el-table-column prop="app_name" :label="t('user.profile.appName')" width="150">
          <template #default="{ row }">
            <el-tag type="info" effect="plain">{{ row.app_name }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="ip" :label="t('user.profile.ip')" width="140" />

        <el-table-column
          prop="user_agent"
          :label="t('user.profile.userAgent')"
          show-overflow-tooltip
        />

        <el-table-column prop="auth_method" :label="t('user.profile.authMethod')" width="110">
          <template #default="{ row }">
            <el-tag size="small" type="warning">{{ row.auth_method }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('user.profile.lastActiveAt')" min-width="100">
          <template #default="scope">
            {{ formatDate(scope.row.last_active_at) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('user.profile.status')" width="110" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_current" type="success" effect="dark" size="small">
              {{ t('user.profile.currentDevice') }}
            </el-tag>
            <el-tag v-else type="info" size="small">{{ t('user.profile.otherDevice') }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column
          :label="t('user.profile.operation')"
          width="100"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              type="danger"
              link
              size="small"
              :disabled="row.is_current"
              @click="revokeSession(toRow(row))"
            >
              {{ t('user.profile.revoke') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<style scoped lang="scss">
.profile-container {
  padding: 20px;
  .mb-4 {
    margin-bottom: 16px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
