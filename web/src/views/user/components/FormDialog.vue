<script setup lang="ts">
import {
  userInfo,
  type UserCreateReq,
  type UserListItemResp,
  type UserUpdateReq,
} from '@/api/modules/user'
import toast, { handleApiError } from '@/utils/message'
import { REGEX_ALPHANUMERIC_UNDERSCORE, REGEX_EMAIL } from '@/utils/validate'
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
          toast.error(data.message)
        }
      } catch (error) {
        handleApiError(error)
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
  username: [
    { required: true, message: t('user.form.usernameRequired'), trigger: 'blur' },
    { min: 1, max: 32, message: t('user.form.usernameLength'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value && !REGEX_ALPHANUMERIC_UNDERSCORE.test(value)) {
          callback(new Error(t('user.form.usernameInvalid')))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  email: [
    { required: true, message: t('user.form.emailRequired'), trigger: 'blur' },
    { min: 6, max: 128, message: t('user.form.emailLength'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value && !REGEX_EMAIL.test(value)) {
          callback(new Error(t('user.form.emailInvalid')))
        } else {
          callback()
        }
      },
      trigger: ['blur', 'change'],
    },
  ],
  password: [
    {
      required: props.mode === 'create',
      message: t('user.form.passwordRequired'),
      trigger: 'blur',
    },
    {
      validator: (_rule, value, callback) => {
        if (props.mode === 'edit' && !value) {
          callback()
          return
        }
        if (value && (value.length < 6 || value.length > 18)) {
          callback(new Error(t('user.form.passwordLength')))
        } else {
          callback()
        }
      },
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
