import { fileURLToPath, URL } from 'node:url'
import path from 'node:path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import IconsResolver from 'unplugin-icons/resolver'
import Icons from 'unplugin-icons/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

const srcPath = fileURLToPath(new URL('./src', import.meta.url))

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),

    AutoImport({
      imports: ['vue', 'vue-router'],
      dts: path.resolve(srcPath, 'auto-imports.d.ts'),
      eslintrc: {
        enabled: true,
        filepath: path.resolve(srcPath, '.eslintrc-auto-import.json'),
        globalsPropValue: true,
      },
      resolvers: [ElementPlusResolver(), IconsResolver({ prefix: 'Icon' })],
    }),
    Components({
      resolvers: [
        ElementPlusResolver(),
        IconsResolver({
          prefix: '',
          enabledCollections: ['ep'],
        }),
      ],
      dts: path.resolve(srcPath, 'components.d.ts'),
    }),
    Icons({
      autoInstall: true,
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
})
