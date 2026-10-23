import request from '../index'
import type { Pagination } from './request'
import type { ApiResponse, ListResp, StatusOption } from './response'

export interface KeypairListReq extends Pagination {
  _brand?: never
}

export interface KeypairListItemResp {
  id: number
  app_id: number
  kid: string
  algorithm: string
  status: string
  activated_at: string
}

export function keypairList(params: KeypairListReq) {
  return request.get<ApiResponse<ListResp<KeypairListItemResp>>>('/keypairs', { params: params })
}

export interface KeypairInfoResp {
  id: number
  app_id: number
  kid: string
  algorithm: string
  public_key: string
  private_key: string
  status: string
  activated_at: string
  retire_at: string
  created_at: string
  updated_at: string
}

export function keypairInfo(id: number) {
  return request.get<ApiResponse<KeypairInfoResp>>(`/keypairs/${id}`)
}

export interface KeypairCreateReq {
  app_id: number
  algorithm: string
  status: string
}

export function keypairCreate(body: KeypairCreateReq) {
  return request.post('/keypairs', body)
}

export function keypairDowngrade(id: number) {
  return request.put<ApiResponse>(`/keypairs/${id}/downgrade`)
}

export function keypairRetire(id: number) {
  return request.put<ApiResponse>(`/keypairs/${id}/retire`)
}

export function keypairStatusOptions() {
  return request.get<ApiResponse<StatusOption[]>>('/keypairs/status-options')
}

export function keypairAlgorithmOptions() {
  return request.get<ApiResponse<StatusOption[]>>('/keypairs/algorithm-options')
}
