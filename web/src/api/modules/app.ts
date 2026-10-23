import request from '../index'
import type { Pagination } from './request'
import type { ApiResponse, ListResp } from './response'

export interface AppListReq extends Pagination {
  keyword?: string
}

export interface AppListItemResp {
  id: number
  name: string
  client_id: string
  status: boolean
  created_at: string
}

export function appList(params: AppListReq) {
  return request.get<ApiResponse<ListResp<AppListItemResp>>>('/apps', { params: params })
}

export interface AppOptionItemResp {
  id: number
  name: string
}

export function appOptions() {
  return request.get<ApiResponse<AppOptionItemResp[]>>('/apps/options')
}

export interface AppInfoResp {
  id: number
  name: string
  client_id: string
  redirect_uris: string[]
  status: boolean
  created_at: string
  updated_at: string
}

export function appInfo(id: number) {
  return request.get<ApiResponse<AppInfoResp>>(`/apps/${id}`)
}

export interface AppCreateReq {
  name: string
  redirect_uris: string[]
  status: boolean
}

export function appCreate(body: AppCreateReq) {
  return request.post<ApiResponse>('/apps', body)
}

export interface AppUpdateReq {
  name: string
  redirect_uris: string[]
  status: boolean
}

export function appUpdate(id: number, body: AppUpdateReq) {
  return request.put<ApiResponse>(`/apps/${id}`, body)
}

export interface AppUpdateStatus {
  status: boolean
}

export function appUpdateStatus(id: number, body: AppUpdateStatus) {
  return request.put<ApiResponse>(`/apps/${id}/status`, body)
}

export function appDelete(id: number) {
  return request.delete<ApiResponse>(`/apps/${id}`)
}
