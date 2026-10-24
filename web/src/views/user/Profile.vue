<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useProfile } from './components/useProfile'
import type { UserSessionItemResp } from '@/api/modules/user'
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { formatDate } from '@/utils/date'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const { loading, userInfo, sessionList, revoke, editPassword, init } = useProfile()
const userStore = useUserStore()
const router = useRouter()

const dialogVisible = ref(false)
const passwordFormRef = ref<FormInstance>()
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const validatePass2 = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (value === '') {
    callback(new Error(t('user.profile.confirmPasswordRequired')))
  } else if (value !== passwordForm.new_password) {
    callback(new Error(t('user.profile.passwordsNotMatch')))
  } else {
    callback()
  }
}

const passwordRules = reactive<FormRules>({
  old_password: [
    { required: true, message: t('user.profile.oldPasswordRequired'), trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: t('user.profile.newPasswordRequired'), trigger: 'blur' },
    { min: 6, message: t('user.profile.passwordMinLength'), trigger: 'blur' },
  ],
  confirm_password: [{ validator: validatePass2, required: true, trigger: 'blur' }],
})

async function handleUpdatePassword() {
  if (!passwordFormRef.value) return
  await passwordFormRef.value.validate(async valid => {
    if (valid) {
      const success = await editPassword({
        old_password: passwordForm.old_password,
        new_password: passwordForm.new_password,
      })
      if (success) {
        dialogVisible.value = false
        passwordFormRef.value?.resetFields()

        userStore.clearMe()

        setTimeout(() => {
          router.push('/login')
        }, 500)
      }
    }
  })
}

function handleDialogClosed() {
  passwordFormRef.value?.resetFields()
}

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
        <div class="card-header flex justify-between items-center">
          <span>{{ t('user.profile.basicInfo') }}</span>
          <el-button type="primary" size="default" @click="dialogVisible = true">
            {{ t('user.profile.changePassword') }}
          </el-button>
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
              @click="revoke(toRow(row))"
            >
              {{ t('user.profile.revoke') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="t('user.profile.changePassword')"
      width="500px"
      @closed="handleDialogClosed"
    >
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="100px"
      >
        <el-form-item :label="t('user.profile.oldPassword')" prop="old_password">
          <el-input v-model="passwordForm.old_password" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('user.profile.newPassword')" prop="new_password">
          <el-input v-model="passwordForm.new_password" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('user.profile.confirmPassword')" prop="confirm_password">
          <el-input v-model="passwordForm.confirm_password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button type="primary" @click="handleUpdatePassword">
            {{ t('user.profile.save') }}
          </el-button>
        </span>
      </template>
    </el-dialog>
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
