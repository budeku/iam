import { appOptions, type AppOptionItemResp } from '@/api/modules/app'
import {
  keypairAlgorithmOptions,
  keypairCreate,
  keypairDowngrade,
  keypairInfo,
  keypairList,
  keypairRetire,
  keypairStatusOptions,
  type KeypairCreateReq,
  type KeypairListItemResp,
  type KeypairListReq,
} from '@/api/modules/keypair'
import type { StatusOption } from '@/api/modules/response'
import { i18n } from '@/locales'
import toast, { handleApiError } from '@/utils/message'
import { ElMessageBox } from 'element-plus'

const { t } = i18n.global

export function useKeypair() {
  const loading = ref(false)
  const submitting = ref(false)
  const listData = ref<KeypairListItemResp[]>([])
  const appOptionData = ref<AppOptionItemResp[]>([])
  const algorithmData = ref<StatusOption[]>([])
  const statusData = ref<StatusOption[]>([])
  const optionsLoaded = ref(false)
  const total = ref(0)
  const params = reactive<KeypairListReq>({
    page: 1,
    per_page: 10,
  })

  async function fetchAppOptions() {
    if (optionsLoaded.value) return
    try {
      const [algoRes, statusRes, appRes] = await Promise.all([
        keypairAlgorithmOptions(),
        keypairStatusOptions(),
        appOptions(),
      ])
      if (algoRes.data.status) algorithmData.value = algoRes.data.data
      if (statusRes.data.status) statusData.value = statusRes.data.data
      if (appRes.data.status) appOptionData.value = appRes.data.data
      optionsLoaded.value = true
    } catch (error) {
      handleApiError(error)
    }
  }

  function getAppName(appId: number) {
    const app = appOptionData.value.find(item => item.id === appId)
    return app ? app.name : appId
  }

  async function query() {
    loading.value = true
    try {
      const { data } = await keypairList(params)
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

  function changePage(page: number) {
    params.page = page
    query()
  }

  function changePageSize(perPage: number) {
    params.per_page = perPage
    params.page = 1
    query()
  }

  async function info(id: number) {
    try {
      const { data } = await keypairInfo(id)
      if (data.status) {
        return data.data
      }
      toast.error(data.message)
      return null
    } catch (error) {
      handleApiError(error)
      return null
    }
  }

  async function add(body: KeypairCreateReq) {
    submitting.value = true
    try {
      const { data } = await keypairCreate(body)
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

  async function downgrade(id: number) {
    try {
      await ElMessageBox.confirm(t('keypair.downgradeConfirm'), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    loading.value = true
    try {
      const { data } = await keypairDowngrade(id)
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

  async function retire(id: number) {
    try {
      await ElMessageBox.confirm(t('keypair.retireConfirm'), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    loading.value = true
    try {
      const { data } = await keypairRetire(id)
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

  async function statusOptions() {
    try {
      const { data } = await keypairStatusOptions()
      if (data.status) {
        statusData.value = data.data
      } else {
        toast.error(data.message)
      }
    } catch (error) {
      handleApiError(error)
    } finally {
      loading.value = false
    }
  }

  async function algorithmOptions() {
    try {
      const { data } = await keypairAlgorithmOptions()
      if (data.status) {
        algorithmData.value = data.data
      } else {
        toast.error(data.message)
      }
    } catch (error) {
      handleApiError(error)
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    submitting,
    params,
    total,
    listData,
    appOptionData,
    algorithmData,
    statusData,
    fetchAppOptions,
    getAppName,
    query,
    changePage,
    changePageSize,
    info,
    add,
    downgrade,
    retire,
    statusOptions,
    algorithmOptions,
  }
}
