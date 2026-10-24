import {
  userMe,
  userSessionRevoke,
  userSessions,
  type UserMeResp,
  type UserSessionItemResp,
} from '@/api/modules/user'
import { i18n } from '@/locales'
import message from '@/utils/message'
import { ElMessageBox } from 'element-plus'

const { t } = i18n.global

export function useProfile() {
  const loading = ref(false)
  const userInfo = ref<UserMeResp | null>(null)
  const sessionList = ref<UserSessionItemResp[]>([])

  async function fetchMe() {
    try {
      const { data } = await userMe()
      if (data.status) {
        userInfo.value = data.data
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    }
  }

  async function fetchSessions() {
    loading.value = true
    try {
      const { data } = await userSessions()
      if (data.status) {
        sessionList.value = data.data
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    } finally {
      loading.value = false
    }
  }

  async function revokeSession(row: UserSessionItemResp) {
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
        message.success(t('user.profile.revokeSuccess'))
        await fetchSessions()
      } else {
        message.error(data.message)
      }
    } catch (error) {
      reportError(error)
    } finally {
      loading.value = false
    }
  }

  async function init() {
    await Promise.all([fetchMe(), fetchSessions()])
  }

  return {
    loading,
    userInfo,
    sessionList,
    fetchMe,
    fetchSessions,
    revokeSession,
    init,
  }
}
