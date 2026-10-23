import { useUserStore } from '@/stores/user'
import type { Router } from 'vue-router'

export function setupAuthGuard(router: Router) {
  router.beforeEach(async to => {
    if (to.meta.public) {
      return true
    }

    const userStore = useUserStore()

    try {
      if (!userStore.me) {
        await userStore.getMe()
      }
    } catch {
      userStore.clearMe()
      return { name: 'login' }
    }

    if (!userStore.me) {
      userStore.clearMe()
      return { name: 'login' }
    }

    const allowedForNonSystem = ['dashboard', 'login', 'NotFound', 'profile']

    if (!userStore.isSuperAdmin && to.name && !allowedForNonSystem.includes(to.name as string)) {
      return { name: 'dashboard' }
    }

    return true
  })
}
