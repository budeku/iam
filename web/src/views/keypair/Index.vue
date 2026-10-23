<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useKeypair } from './components/useKeypair'
import type { KeypairCreateReq, KeypairListItemResp } from '@/api/modules/keypair'
import { formatDate } from '@/utils/date'
import FormDialog from './components/FormDialog.vue'
import DetailDialog from './components/DetailDialog.vue'

const { t } = useI18n()

const {
  loading,
  submitting,
  params,
  total,
  listData,
  query,
  changePage,
  changePageSize,
  add,
  downgrade,
  retire,
} = useKeypair()

function toRow(row: unknown): KeypairListItemResp {
  return row as KeypairListItemResp
}

const dialogVisible = ref(false)
const detailVisible = ref(false)
const activeRow = ref<KeypairListItemResp | null>(null)

function openAdd() {
  dialogVisible.value = true
}

function openDetail(row: KeypairListItemResp) {
  activeRow.value = row
  detailVisible.value = true
}

async function handleFormSubmit(payload: KeypairCreateReq) {
  const ok = await add(payload)
  if (ok) {
    dialogVisible.value = false
  }
}

query()
</script>

<template>
  <div class="keypair-list-view">
    <el-card shadow="never">
      <div class="toolbar">
        <el-button type="primary" @click="openAdd">
          {{ t('keypair.form.createTitle') }}
        </el-button>
      </div>

      <el-table
        border
        v-loading="loading"
        :data="listData"
        :header-cell-style="{ textAlign: 'center' }"
        :cell-style="{ textAlign: 'center' }"
        style="width: 100%"
      >
        <el-table-column
          type="index"
          :index="i => (params.page! - 1) * params.per_page! + i + 1"
          :label="t('common.recordNumber')"
          width="70"
        />

        <el-table-column :label="t('keypair.field.appName')" prop="app_name" min-width="100" />
        <el-table-column :label="t('keypair.field.kid')" prop="kid" min-width="200" />
        <el-table-column :label="t('keypair.field.algorithm')" min-width="100">
          <template #default="scope">
            <el-tag type="primary">{{ toRow(scope.row).algorithm }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('keypair.field.status')" min-width="100">
          <template #default="scope">
            <el-tag type="success">{{ toRow(scope.row).status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('keypair.field.activatedAt')" min-width="160">
          <template #default="scope">
            {{ formatDate(scope.row.activated_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="280" fixed="right" align="left">
          <template #default="scope">
            <el-button type="primary" link @click="openDetail(toRow(scope.row))">
              {{ t('common.detail') }}
            </el-button>

            <template v-if="scope.row.app_id !== 0">
              <el-button
                v-if="scope.row.status === 'active'"
                type="primary"
                link
                @click="downgrade(toRow(scope.row).id)"
              >
                {{ t('keypair.downgrade') }}
              </el-button>

              <el-button
                v-else-if="scope.row.status === 'grace'"
                type="danger"
                link
                @click="retire(toRow(scope.row).id)"
              >
                {{ t('keypair.retire') }}
              </el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        layout="total, sizes, prev, pager, next"
        :page-sizes="[10, 20, 50]"
        :current-page="params.page"
        :page-size="params.per_page"
        :total="total"
        @current-change="changePage"
        @size-change="changePageSize"
      />
    </el-card>

    <FormDialog
      v-model:visible="dialogVisible"
      v-model:submitting="submitting"
      @submit="handleFormSubmit"
    />

    <DetailDialog v-model:visible="detailVisible" :row="activeRow" />
  </div>
</template>

<style scoped lang="scss">
.keypair-list-view {
  padding: 20px;

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .toolbar-left {
      display: flex;
      gap: 8px;
    }
  }

  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
