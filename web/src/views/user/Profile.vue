<script setup lang="ts">
import { onMounted } from 'vue'
import { useProfile } from './components/useProfile'
import type { UserSessionItemResp } from '@/api/modules/user'
import { useI18n } from 'vue-i18n'

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
    <!-- 1. 基本信息卡片 -->
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

    <!-- 2. 会话与设备管理卡片 -->
    <el-card class="box-card" shadow="hover">
      <template #header>
        <div class="card-header flex justify-between items-center">
          <span>{{ t('user.profile.sessionManagement') }}</span>
          <el-button size="small" @click="init">{{ t('user.profile.refresh') }}</el-button>
        </div>
      </template>

      <el-table :data="sessionList" style="width: 100%" border stripe>
        <!-- 展开行查看内部关联的 RefreshToken -->
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="p-4 bg-gray-50">
              <p class="font-bold text-sm mb-2 text-gray-600">
                {{ t('user.profile.expandRefreshTokens') }}
              </p>
              <el-table :data="row.refresh_tokens" size="small" border>
                <el-table-column prop="id" :label="t('user.profile.tokenId')" width="90" />
                <el-table-column prop="jti" :label="t('user.profile.jti')" />
                <el-table-column
                  prop="expires_at"
                  :label="t('user.profile.expiresAt')"
                  width="180"
                />
              </el-table>
            </div>
          </template>
        </el-table-column>

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

        <el-table-column
          prop="last_active_at"
          :label="t('user.profile.lastActiveAt')"
          width="170"
        />

        <el-table-column :label="t('user.profile.status')" width="110" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_current" type="success" effect="dark" size="small">
              {{ t('user.profile.currentDevice') }}
            </el-tag>
            <el-tag v-else type="info" size="small">{{ t('user.profile.otherDevice') }}</el-tag>
          </template>
        </el-table-column>

        <!-- 操作列 -->
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
}
</style>
