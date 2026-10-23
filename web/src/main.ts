import { createApp } from 'vue'
import pinia from './stores/pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import App from './App.vue'
import router from './router'
import './styles/index.scss'
import { i18n } from '@/locales'

const app = createApp(App)

app.use(pinia)
app.use(i18n)
app.use(ElementPlus)
app.use(router)

app.mount('#app')
