import { appConfig } from '@/configs'
import { getItem } from '@/utils/storage'
import { createI18n } from 'vue-i18n'
import zhCN from './lang/zh-CN'

export const SUPPORTED_LOCALES = [
  { value: 'zh-CN', label: '简体中文' },
  // 以后在这里扩展其他语言
] as const

export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number]['value']

const messages = {
  'zh-CN': zhCN,
}

function isSupportedLocale(value: unknown): value is SupportedLocale {
  return SUPPORTED_LOCALES.some(item => item.value === value)
}

function resolveInitialLocale(): SupportedLocale {
  const storedLocale = getItem<string>(appConfig.localeStorageKey, 'cookie')
  if (isSupportedLocale(storedLocale)) {
    return storedLocale
  }

  if (typeof navigator !== 'undefined' && navigator.language) {
    const browserLang = navigator.language
    const matched = SUPPORTED_LOCALES.find(item =>
      browserLang.startsWith(item.value.split('-')[0] ?? '')
    )
    if (matched) {
      return matched.value
    }
  }

  return 'zh-CN'
}

export const i18n = createI18n({
  legacy: false,
  locale: resolveInitialLocale(),
  fallbackLocale: 'zh-CN',
  messages,
})
