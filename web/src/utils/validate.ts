// 邮箱正则
export const REGEX_EMAIL = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/

export function isValidEmail(email: string): boolean {
  return REGEX_EMAIL.test(email)
}

// 仅包含数字、字母、下划线
export const REGEX_ALPHANUMERIC_UNDERSCORE = /^[a-zA-Z0-9_]+$/

export function isValidAlphanumericUnderscore(value: string): boolean {
  return REGEX_ALPHANUMERIC_UNDERSCORE.test(value)
}
