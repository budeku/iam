<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { KeypairListItemResp, KeypairInfoResp } from '@/api/modules/keypair'
import { useKeypair } from './useKeypair'
import { formatDate } from '@/utils/date'

const { t } = useI18n()
const { appOptionData, info } = useKeypair()

const props = defineProps<{
  row?: KeypairListItemResp | null
}>()

const visible = defineModel<boolean>('visible', { default: false })
const form = ref<Partial<KeypairInfoResp>>({})

watch(
  visible,
  async vis => {
    if (vis && props.row) {
      const data = await info(props.row.id)
      if (data) {
        form.value = data
      }
    }
  },
  { immediate: true }
)
</script>

<template>
  <el-dialog v-model="visible" :title="t('keypair.form.detailTitle')" width="800px">
    <el-form :model="form" label-width="110px" disabled>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="ID">
            <el-input v-model.number="form.id" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('keypair.field.appId')">
            <el-select v-model="form.app_id" style="width: 100%">
              <el-option
                v-for="item in appOptionData"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="t('keypair.field.kid')">
            <el-input v-model="form.kid" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('keypair.field.algorithm')">
            <el-input :model-value="form.algorithm" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="t('keypair.field.status')">
            <el-input :model-value="form.status" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('keypair.field.createdAt')">
            <el-input :model-value="formatDate(form.created_at!)" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item :label="t('keypair.field.activatedAt')">
            <el-input :model-value="formatDate(form.activated_at!)" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('keypair.field.retireAt')">
            <el-input :model-value="formatDate(form.retire_at!)" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item :label="t('keypair.field.publicKey')">
        <el-input v-model="form.public_key" type="textarea" :rows="3" />
      </el-form-item>

      <el-form-item :label="t('keypair.field.privateKey')">
        <el-input v-model="form.private_key" type="textarea" :rows="3" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button type="primary" @click="visible = false">{{ t('common.close') }}</el-button>
    </template>
  </el-dialog>
</template>
