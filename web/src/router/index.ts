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
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/dashboard/Index.vue'),
          meta: { icon: 'monitor', name: 'dashboard' },
        },
        {
          path: 'apps',
          name: 'apps',
          component: () => import('@/views/app/Index.vue'),
          meta: { icon: 'grid', name: 'app' },
        },
        {
          path: 'keypairs',
          name: 'keypairs',
          component: () => import('@/views/keypair/Index.vue'),
          meta: { icon: 'key', name: 'keypair' },
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('@/views/user/Profile.vue'),
          meta: { icon: 'user', name: 'profile', hidden: true },
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
