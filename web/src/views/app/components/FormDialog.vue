<script setup lang="ts">
import {
  appInfo,
  type AppCreateReq,
  type AppListItemResp,
  type AppUpdateReq,
} from '@/api/modules/app'
import { message, reportError } from '@/utils/message'
import type { FormInstance, FormRules } from 'element-plus'
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  mode: 'create' | 'edit'
  row?: AppListItemResp | null
}>()

const visible = defineModel<boolean>('visible', { default: false })
const submitting = defineModel<boolean>('submitting', { default: false })

const emit = defineEmits<{
  submit: [payload: AppCreateReq | AppUpdateReq, id?: number]
}>()

const formRef = ref<FormInstance>()

const form = reactive({
  name: '',
  redirect_uris: '',
  status: true,
})

function resetForm() {
  form.name = ''
  form.redirect_uris = ''
  form.status = true
  formRef.value?.clearValidate()
}

watch(
  () => props.row,
  async row => {
    if (props.mode === 'edit' && row) {
      try {
        const { data } = await appInfo(row.id)
        if (data.status) {
          form.name = data.data.name
          form.redirect_uris = Array.isArray(data.data.redirect_uris)
            ? data.data.redirect_uris.join('\n')
            : data.data.redirect_uris || ''
          form.status = data.data.status // 把接口返回的布尔值赋值给表单
        } else {
          message.error(data.message)
        }
      } catch (error) {
        reportError(error)
      }
    }
  },
  { immediate: true }
)

const rules: FormRules = {
  name: [{ required: true, message: t('app.form.nameRequired'), trigger: 'blur' }],
  redirect_uris: [{ required: true, message: t('app.form.redirectUrisRequired'), trigger: 'blur' }],
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  const uris = form.redirect_uris
    .split('\n')
    .map(s => s.trim())
    .filter(Boolean)

  if (props.mode === 'create') {
    emit('submit', {
      name: form.name,
      redirect_uris: uris,
    })
  } else {
    emit(
      'submit',
      {
        name: form.name,
        status: form.status,
        redirect_uris: uris,
      },
      props.row?.id
    )
  }
}

function handleClosed() {
  resetForm()
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="mode === 'create' ? t('app.form.createTitle') : t('app.form.editTitle')"
    width="520px"
    @closed="handleClosed"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item :label="t('app.field.name')" prop="name">
        <el-input v-model="form.name" :placeholder="t('app.form.namePlaceholder')" />
      </el-form-item>

      <el-form-item :label="t('app.field.redirectUris')" prop="redirect_uris">
        <el-input
          v-model="form.redirect_uris"
          type="textarea"
          :rows="3"
          :placeholder="t('app.form.redirectUrisPlaceholder')"
        />
      </el-form-item>

      <el-form-item v-if="mode === 'edit'" :label="t('app.field.status')" prop="status">
        <el-select v-model="form.status" style="width: 100%">
          <el-option :label="t('app.status.active')" :value="true" />
          <el-option :label="t('app.status.disabled')" :value="false" />
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
