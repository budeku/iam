<script setup lang="ts">
import {
  userInfo,
  type UserCreateReq,
  type UserListItemResp,
  type UserUpdateReq,
} from '@/api/modules/user'
import message from '@/utils/message'
import type { FormInstance, FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  mode: 'create' | 'edit'
  row?: UserListItemResp | null
}>()

const visible = defineModel<boolean>('visible', { default: false })
const submitting = defineModel<boolean>('submitting', { default: false })

const emit = defineEmits<{
  submit: [payload: UserCreateReq | UserUpdateReq, id?: number]
}>()

const formRef = ref<FormInstance>()

const form = reactive({
  username: '',
  email: '',
  password: '',
  status: true,
})

function resetForm() {
  form.username = ''
  form.email = ''
  form.password = ''
  form.status = true
  formRef.value?.clearValidate()
}

watch(
  () => props.row,
  async row => {
    if (props.mode === 'edit' && row) {
      try {
        const { data } = await userInfo(row.id)
        if (data.status) {
          form.username = data.data.username
          form.email = data.data.email
          form.status = data.data.status
          form.password = ''
        } else {
          message.error(data.message)
        }
      } catch (error) {
        reportError(error)
      }
    } else {
      resetForm()
    }

    await nextTick()
    formRef.value?.clearValidate()
  },
  { immediate: true }
)

watch(visible, val => {
  if (val) {
    nextTick(() => {
      formRef.value?.clearValidate()
    })
  }
})

const rules = computed<FormRules>(() => ({
  username: [{ required: true, message: t('user.form.usernameRequired'), trigger: 'blur' }],
  email: [
    { required: true, message: t('user.form.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('user.form.emailInvalid'), trigger: ['blur', 'change'] },
  ],
  password: [
    {
      required: props.mode === 'create',
      message: t('user.form.passwordRequired'),
      trigger: 'blur',
    },
  ],
}))

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  if (props.mode === 'create') {
    const payload: UserCreateReq = {
      username: form.username,
      email: form.email,
      password: form.password,
    }
    emit('submit', payload)
  } else {
    const payload: UserUpdateReq = {
      username: form.username,
      email: form.email,
      status: form.status,
    }
    // 密码留空不修改
    if (form.password) {
      payload.password = form.password
    }
    emit('submit', payload, props.row?.id)
  }
}

function handleClosed() {
  resetForm()
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="mode === 'create' ? t('user.form.createTitle') : t('user.form.editTitle')"
    width="520px"
    @closed="handleClosed"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item :label="t('user.field.username')" prop="username">
        <el-input v-model="form.username" :placeholder="t('user.form.usernameRequired')" />
      </el-form-item>

      <el-form-item :label="t('user.field.email')" prop="email">
        <el-input v-model="form.email" :placeholder="t('user.form.emailRequired')" />
      </el-form-item>

      <el-form-item :label="t('user.field.password')" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          :placeholder="
            mode === 'edit' ? t('user.form.passwordEditHint') : t('user.form.passwordRequired')
          "
        />
      </el-form-item>

      <el-form-item v-if="mode === 'edit'" :label="t('user.field.status')" prop="status">
        <el-select v-model="form.status" style="width: 100%">
          <el-option :label="t('user.status.active')" :value="true" />
          <el-option :label="t('user.status.disabled')" :value="false" />
        </el-select>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>
