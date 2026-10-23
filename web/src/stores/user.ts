import { userMe, type UserMeResp } from '@/api/modules/user'
import { defineStore } from 'pinia'
import { ref as vueRef } from 'vue'

export const useUserStore = defineStore('user', () => {
  const me = vueRef<UserMeResp | null>(null)
  const loading = vueRef(false)

  const isSuperAdmin = computed(() => me.value?.is_super_admin ?? false)

  async function getMe() {
    if (me.value) return me.value

    try {
      loading.value = true
      const { data } = await userMe()
      if (data.status) {
        me.value = data.data
      }
    } catch (error) {
      console.error(error)
    } finally {
      loading.value = false
    }
  }

  function clearMe() {
    me.value = null
  }

  return {
    me,
    loading,
    isSuperAdmin,
    getMe,
    clearMe,
  }
})
