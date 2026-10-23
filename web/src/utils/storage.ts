export type StorageType = 'local' | 'session' | 'cookie'

const PREFIX = import.meta.env.VITE_STORAGE_PREFIX ?? 'iam-web'

const withPrefix = (key: string): string => `${PREFIX}_${key}`

export function safeParse<T>(value: string | null): T | null {
  if (!value) return null
  try {
    return JSON.parse(value) as T
  } catch {
    return null
  }
}

export function safeStringify(value: unknown): string {
  try {
    return JSON.stringify(value)
  } catch {
    return ''
  }
}

interface StorageLike {
  getItem: (key: string) => string | null
  setItem: (key: string, value: string) => void
  removeItem: (key: string) => void
  clear: () => void
}

const cookieStorage: StorageLike = {
  getItem: key => {
    if (typeof document === 'undefined') return null
    const match = document.cookie.match(
      new RegExp('(?:^|; )' + key.replace(/([.$?*|{}()[\]\\/+^:])/g, '\\$1') + '=([^;]*)')
    )
    return match ? decodeURIComponent(match[1] ?? '') : null
  },
  setItem: (key, value, days = 365) => {
    if (typeof document === 'undefined') return
    const expires = new Date(Date.now() + days * 24 * 60 * 60 * 1000).toUTCString()
    document.cookie = `${key}=${encodeURIComponent(value)}; expires=${expires}; path=/; SameSite=Lax`
  },
  removeItem: key => {
    if (typeof document === 'undefined') return
    document.cookie = `${key}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`
  },
  clear: () => {
    if (typeof document === 'undefined') return
    document.cookie
      .split('; ')
      .map(item => item.split('=')[0])
      .filter((key): key is string => Boolean(key))
      .forEach(key => cookieStorage.removeItem(key))
  },
}

const getNativeStorage = (type: StorageType): StorageLike => {
  if (type === 'cookie') return cookieStorage
  // 兼顾 SSR 环境，避免在服务端直接调用 localStorage/sessionStorage 报错
  if (typeof window === 'undefined') {
    return { getItem: () => null, setItem: () => {}, removeItem: () => {}, clear: () => {} }
  }
  return type === 'session' ? sessionStorage : localStorage
}

// API
const getDefaultStorageType = (): StorageType => {
  const env = import.meta.env?.VITE_MENU_STORAGE
  if (env === 'session' || env === 'cookie') return env
  return 'local'
}

export function setItem<T>(
  key: string,
  value: T,
  type: StorageType = getDefaultStorageType()
): void {
  const storage = getNativeStorage(type)
  storage.setItem(withPrefix(key), safeStringify(value))
}

export function getItem<T>(key: string, type: StorageType = getDefaultStorageType()): T | null {
  const storage = getNativeStorage(type)
  const raw = storage.getItem(withPrefix(key))
  return safeParse<T>(raw)
}

export function getItemWithDefault<T>(key: string, defaultValue: T, type?: StorageType): T {
  const value = getItem<T>(key, type)
  return value ?? defaultValue
}

export function removeItem(key: string, type: StorageType = getDefaultStorageType()): void {
  const storage = getNativeStorage(type)
  storage.removeItem(withPrefix(key))
}

export function clear(type: StorageType = getDefaultStorageType()): void {
  const storage = getNativeStorage(type)
  storage.clear()
}
