<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useUser } from './components/useUser'
import type { UserCreateReq, UserListItemResp, UserUpdateReq } from '@/api/modules/user'
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
} = useUser()

function toRow(row: unknown): UserListItemResp {
  return row as UserListItemResp
}

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const activeRow = ref<UserListItemResp | null>(null)

function openAdd() {
  dialogMode.value = 'create'
  activeRow.value = null
  dialogVisible.value = true
}

function openEdit(row: UserListItemResp) {
  dialogMode.value = 'edit'
  activeRow.value = row
  dialogVisible.value = true
}

async function handleFormSubmit(payload: UserCreateReq | UserUpdateReq, id?: number) {
  const ok =
    dialogMode.value === 'create'
      ? await add(payload as UserCreateReq)
      : await edit(id!, payload as UserUpdateReq)
  if (ok) dialogVisible.value = false
}

async function handleStatusChange(row: UserListItemResp) {
  const ok = await editStatus(row.id, !row.status)
  if (!ok) {
    query()
  }
}

query()
</script>

<template>
  <div class="app-list-view">
    <el-card shadow="never">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="params.keyword"
            :placeholder="t('user.searchPlaceholder')"
            clearable
            :maxlength="26"
            style="width: 230px"
            @keyup.enter="search"
          />
          <el-button type="primary" @click="search">{{ t('common.search') }}</el-button>
          <el-button @click="reset">{{ t('common.reset') }}</el-button>
        </div>
        <el-button type="primary" @click="openAdd">
          {{ t('user.form.createTitle') }}
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
        <el-table-column :label="t('user.field.username')" prop="username" min-width="140" />
        <el-table-column :label="t('user.field.email')" prop="email" min-width="160" />
        <el-table-column :label="t('user.field.uuid')" prop="uuid" min-width="160">
          <template #default="scope">
            <el-tooltip :content="scope.row.uuid" placement="top" effect="dark">
              <div class="table-cell-ellipsis">
                {{ scope.row.uuid }}
              </div>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column :label="t('user.field.status')" width="120">
          <template #default="scope">
            <el-switch
              inline-prompt
              :model-value="scope.row.status"
              :active-text="t('user.status.active')"
              :inactive-text="t('user.status.disabled')"
              :disabled="scope.row.is_super_admin"
              @change="val => handleStatusChange(toRow(scope.row))"
            />
          </template>
        </el-table-column>
        <el-table-column :label="t('user.field.LastLoginAt')" min-width="160">
          <template #default="scope">
            {{ formatDate(scope.row.last_login_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" width="280" fixed="right" align="left">
          <template #default="scope">
            <el-tooltip
              :content="scope.row.is_super_admin ? t('user.form.superAdminTip') : ''"
              placement="top"
              :disabled="!scope.row.is_super_admin"
            >
              <span>
                <el-button
                  type="primary"
                  link
                  :disabled="scope.row.is_super_admin"
                  @click="openEdit(toRow(scope.row))"
                >
                  {{ t('common.edit') }}
                </el-button>
              </span>
            </el-tooltip>

            <el-tooltip
              :content="scope.row.is_super_admin ? t('user.form.superAdminTip') : ''"
              placement="top"
              :disabled="!scope.row.is_super_admin"
            >
              <span>
                <el-button
                  type="danger"
                  link
                  :disabled="scope.row.is_super_admin"
                  @click="remove(toRow(scope.row))"
                >
                  {{ t('common.delete') }}
                </el-button>
              </span>
            </el-tooltip>
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
.app-list-view {
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

  .table-cell-ellipsis {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 100%;
  }

  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
