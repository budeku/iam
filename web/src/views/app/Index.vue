<script setup lang="ts">
import type { AppCreateReq, AppListItemResp, AppUpdateReq } from '@/api/modules/app'
import { useI18n } from 'vue-i18n'
import { useApp } from './components/useApp'
import { formatDate } from '@/utils/date'
import FormDialog from './components/FormDialog.vue'

const { t } = useI18n()

const {
  loading,
  submitting,
  params,
  total,
  listData,
  query,
  reset,
  search,
  changePage,
  changePageSize,
  add,
  edit,
  editStatus,
  remove,
} = useApp()

function toRow(row: unknown): AppListItemResp {
  return row as AppListItemResp
}

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const activeRow = ref<AppListItemResp | null>(null)

function openAdd() {
  dialogMode.value = 'create'
  activeRow.value = null
  dialogVisible.value = true
}

function openEdit(row: AppListItemResp) {
  dialogMode.value = 'edit'
  activeRow.value = row
  dialogVisible.value = true
}

async function handleFormSubmit(payload: AppCreateReq | AppUpdateReq, id?: number) {
  const ok =
    dialogMode.value === 'create'
      ? await add(payload as AppCreateReq)
      : await edit(id!, payload as AppUpdateReq)
  if (ok) dialogVisible.value = false
}

async function handleStatusChange(row: AppListItemResp) {
  const ok = await editStatus(row.id, !row.status)
  if (!ok) {
    query()
  }
}

query()
</script>

<template>
  <div class="application-list-view">
    <el-card shadow="never">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="params.keyword"
            :placeholder="t('app.searchPlaceholder')"
            clearable
            style="width: 220px"
            @keyup.enter="search"
          />
          <el-button type="primary" @click="search">{{ t('common.search') }}</el-button>
          <el-button @click="reset">{{ t('common.reset') }}</el-button>
        </div>
        <el-button type="primary" @click="openAdd">
          {{ t('app.form.createTitle') }}
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
        <el-table-column :label="t('app.field.name')" prop="name" min-width="140" />
        <el-table-column :label="t('app.field.clientId')" prop="client_id" min-width="160" />
        <el-table-column :label="t('app.field.status')" width="120">
          <template #default="scope">
            <el-switch
              inline-prompt
              :model-value="scope.row.status"
              :active-text="t('app.status.active')"
              :inactive-text="t('app.status.disabled')"
              @change="val => handleStatusChange(toRow(scope.row))"
            />
          </template>
        </el-table-column>
        <el-table-column :label="t('app.field.createdAt')" min-width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="280" fixed="right" align="left">
          <template #default="scope">
            <el-button type="primary" link @click="openEdit(toRow(scope.row))">
              {{ t('common.edit') }}
            </el-button>
            <el-button type="danger" link @click="remove(toRow(scope.row))">
              {{ t('common.delete') }}
            </el-button>
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
      :mode="dialogMode"
      :row="activeRow"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<style scoped lang="scss">
.application-list-view {
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
