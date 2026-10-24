import {
  appCreate,
  appDelete,
  appList,
  appUpdate,
  appUpdateStatus,
  type AppCreateReq,
  type AppListItemResp,
  type AppListReq,
  type AppUpdateReq,
} from '@/api/modules/app'
import { i18n } from '@/locales'
import toast, { handleApiError } from '@/utils/message'
import { ElMessageBox } from 'element-plus'

const { t } = i18n.global

export function useApp() {
  const loading = ref(false)
  const submitting = ref(false)
  const listData = ref<AppListItemResp[]>([])
  const total = ref(0)
  const params = reactive<AppListReq>({
    page: 1,
    per_page: 10,
    keyword: '',
  })

  async function query() {
    loading.value = true
    try {
      const { data } = await appList(params)
      if (data.status) {
        listData.value = data.data.content
        total.value = data.data.count
      } else {
        toast.error(data.message)
      }
    } catch (error) {
      handleApiError(error)
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

  async function add(body: AppCreateReq) {
    submitting.value = true
    try {
      const { data } = await appCreate(body)
      if (data.status) {
        toast.success()
        await query()
        return true
      }
      toast.error(data.message)
      return false
    } catch (error) {
      handleApiError(error)
      return false
    } finally {
      submitting.value = false
    }
  }

  async function edit(id: number, body: AppUpdateReq) {
    submitting.value = true
    try {
      const { data } = await appUpdate(id, body)
      if (data.status) {
        toast.success()
        await query()
        return true
      }
      toast.error(data.message)
      return false
    } catch (error) {
      handleApiError(error)
      return false
    } finally {
      submitting.value = false
    }
  }

  async function remove(row: AppListItemResp) {
    try {
      await ElMessageBox.confirm(t('app.deleteConfirm', { name: row.name }), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    loading.value = true
    try {
      const { data } = await appDelete(row.id)
      if (data.status) {
        toast.success()
        if (listData.value.length === 1 && params.page! > 1) {
          params.page!--
        }
        await query()
      } else {
        toast.error(data.message)
      }
    } catch (error) {
      handleApiError(error)
    } finally {
      loading.value = false
    }
  }

  async function editStatus(id: number, status: boolean) {
    loading.value = true
    try {
      const { data } = await appUpdateStatus(id, { status })
      if (data.status) {
        toast.success()
        await query()
        return true
      }
      toast.error(data.message)
      return false
    } catch (error) {
      handleApiError(error)
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    submitting,
    listData,
    total,
    params,
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
