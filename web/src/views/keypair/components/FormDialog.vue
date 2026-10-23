<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import type { KeypairCreateReq } from '@/api/modules/keypair'
import { useKeypair } from './useKeypair'

const { t } = useI18n()
const { appOptionData, algorithmData, statusData, fetchAppOptions } = useKeypair()

const visible = defineModel<boolean>('visible', { default: false })
const submitting = defineModel<boolean>('submitting', { default: false })

const emit = defineEmits<{
  submit: [payload: KeypairCreateReq]
}>()

const formRef = ref<FormInstance>()
const form = reactive<KeypairCreateReq>({
  app_id: 1,
  algorithm: '',
  status: '',
})

onMounted(() => {
  fetchAppOptions()
})

function resetForm() {
  form.app_id = 1
  form.algorithm = ''
  form.status = ''
  formRef.value?.clearValidate()
}

const rules: FormRules = {
  app_id: [{ required: true, message: t('keypair.form.appIdRequired'), trigger: 'change' }],
  algorithm: [{ required: true, message: t('keypair.form.algorithmRequired'), trigger: 'change' }],
  status: [{ required: true, message: t('keypair.form.statusRequired'), trigger: 'change' }],
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  emit('submit', { ...form })
}

function handleClosed() {
  resetForm()
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="t('keypair.form.createTitle')"
    width="520px"
    @closed="handleClosed"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
      <el-form-item :label="t('keypair.field.appId')" prop="app_id">
        <el-select v-model="form.app_id" style="width: 100%">
          <el-option
            v-for="item in appOptionData"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('keypair.field.algorithm')" prop="algorithm">
        <el-select v-model="form.algorithm" style="width: 100%">
          <el-option
            v-for="item in algorithmData"
            :key="item.label"
            :label="item.label"
            :value="item.label"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('keypair.field.status')" prop="status">
        <el-select v-model="form.status" style="width: 100%">
          <el-option
            v-for="item in statusData"
            :key="item.label"
            :label="item.label"
            :value="item.label"
          />
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
