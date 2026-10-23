import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/layouts/Index.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/login/Index.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: Layout,
      redirect: '/dashboard',
      children: [
        {
          path: '/dashboard',
          name: 'dashboard',
          component: () => import('@/views/dashboard/Index.vue'),
          meta: { icon: 'monitor', name: 'dashboard' },
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/login/Index.vue'),
      meta: { public: true },
    },
  ],
})

export default router
