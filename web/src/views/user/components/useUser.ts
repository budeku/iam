import {
  userCreate,
  userDelete,
  userList,
  userUpdate,
  userUpdateStatus,
  type UserCreateReq,
  type UserListItemResp,
  type UserListReq,
  type UserUpdateReq,
} from '@/api/modules/user'
import { i18n } from '@/locales'
import message from '@/utils/message'
import { ElMessageBox } from 'element-plus'

const { t } = i18n.global

export function useUser() {
  const loading = ref(false)
  const submitting = ref(false)
  const listData = ref<UserListItemResp[]>([])
  const total = ref(0)
  const params = reactive<UserListReq>({
    page: 1,
    per_page: 10,
    keyword: '',
  })

  async function query() {
    loading.value = true
    try {
      const { data } = await userList(params)
      if (data.status) {
        listData.value = data.data.content
        total.value = data.data.count
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    } finally {
      loading.value = false
    }
  }

  function search() {
    params.page = 1
    query()
  }

  function reset() {
    params.page = 1
    params.keyword = ''
    query()
  }

  function changePage(page: number) {
    params.page = page
    query()
  }

  function changePageSize(perPage: number) {
    params.per_page = perPage
    params.page = 1
    query()
  }

  async function add(body: UserCreateReq) {
    submitting.value = true
    try {
      const { data } = await userCreate(body)
      if (data.status) {
        message.success()
        await query()
        return true
      }
      message.error(data.message)
      return false
    } catch (error) {
      reportError(error)
      return false
    } finally {
      submitting.value = false
    }
  }

  async function edit(id: number, body: UserUpdateReq) {
    submitting.value = true
    try {
      const { data } = await userUpdate(id, body)
      if (data.status) {
        message.success()
        await query()
        return true
      }
      message.error(data.message)
      return false
    } catch (error) {
      reportError(error)
      return false
    } finally {
      submitting.value = false
    }
  }

  async function remove(row: UserListItemResp) {
    try {
      await ElMessageBox.confirm(t('app.deleteConfirm', { name: row.username }), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    loading.value = true
    try {
      const { data } = await userDelete(row.id)
      if (data.status) {
        message.success()
        if (listData.value.length === 1 && params.page! > 1) {
          params.page!--
        }
        await query()
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    } finally {
      loading.value = false
    }
  }

  async function editStatus(id: number, status: boolean) {
    loading.value = true
    try {
      const { data } = await userUpdateStatus(id, { status })
      if (data.status) {
        message.success()
        await query()
        return true
      }
      message.error(data.message)
      return false
    } catch (error) {
      reportError(error)
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    submitting,
    params,
    listData,
    total,
    query,
    search,
    reset,
    changePage,
    changePageSize,
    add,
    edit,
    editStatus,
    remove,
  }
}
