export interface ApiResponse<T = unknown> {
  code: number
  status: boolean
  message: string
  data: T
}

export interface ListResp<T = unknown> {
  content: T[]
  count: number
}

export interface StatusOption {
  value: number
  label: string
}
