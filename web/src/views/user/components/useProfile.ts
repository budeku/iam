import {
  userMe,
  userSessionRevoke,
  userSessions,
  userUpdatePassword,
  type UserMeResp,
  type UserSessionItemResp,
  type UserUpdatePasswordReq,
} from '@/api/modules/user'
import { i18n } from '@/locales'
import toast, { handleApiError } from '@/utils/message'
import { ElMessageBox } from 'element-plus'

const { t } = i18n.global

export function useProfile() {
  const loading = ref(false)
  const userInfo = ref<UserMeResp | null>(null)
  const sessionList = ref<UserSessionItemResp[]>([])

  async function me() {
    try {
      const { data } = await userMe()
      if (data.status) {
        userInfo.value = data.data
      } else {
        toast.error(data.message)
      }
    } catch (error) {
      handleApiError(error)
    }
  }

  async function sessions() {
    loading.value = true
    try {
      const { data } = await userSessions()
      if (data.status) {
        sessionList.value = data.data
      } else {
        toast.error(data.message)
      }
    } catch (error) {
      handleApiError(error)
    } finally {
      loading.value = false
    }
  }

  async function revoke(row: UserSessionItemResp) {
    try {
      await ElMessageBox.confirm(t('user.profile.revokeConfirm', { ip: row.ip }), t('common.tip'), {
        type: 'warning',
      })
    } catch {
      return
    }

    loading.value = true
    try {
      const { data } = await userSessionRevoke(row.id)
      if (data.status) {
        toast.success(t('user.profile.revokeSuccess'))
        await sessions()
      } else {
        toast.error(data.message)
      }
    } catch (error) {
      handleApiError(error)
    } finally {
      loading.value = false
    }
  }

  async function editPassword(payload: UserUpdatePasswordReq) {
    loading.value = true
    try {
      const { data } = await userUpdatePassword(payload)
      if (data.status) {
        toast.success(t('user.profile.updatePasswordSuccess'))
        return true
      } else {
        toast.error(data.message)
        return false
      }
    } catch (error) {
      handleApiError(error)
      return false
    } finally {
      loading.value = false
    }
  }

  async function init() {
    await Promise.all([me(), sessions()])
  }

  return {
    loading,
    userInfo,
    sessionList,
    me,
    sessions,
    revoke,
    editPassword,
    init,
  }
}
