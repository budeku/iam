import request from '../index'
import type { Pagination } from './request'
import type { ApiResponse, ListResp } from './response'

export interface UserLoginReq {
  username: string
  password: string
}

export function login(body: UserLoginReq) {
  return request.post('/auth/login', body)
}

export function logout() {
  return request.post<ApiResponse>('/auth/logout')
}

export interface UserMeResp {
  username: string
  email: string
  uuid: string
  is_super_admin: boolean
}

export function userMe() {
  return request.get<ApiResponse<UserMeResp>>('/auth/me')
}

export interface UserListReq extends Pagination {
  keyword?: string
}

export interface UserListItemResp {
  id: number
  username: string
  email: string
  uuid: string
  status: boolean
  last_login_at: string
  is_super_admin: boolean
}

export function userList(params: UserListReq) {
  return request.get<ApiResponse<ListResp<UserListItemResp>>>('/users', { params: params })
}

export interface UserInfoResp {
  id: number
  username: string
  uuid: string
  email: string
  status: boolean
  last_login_at: string
  is_super_admin: boolean
}

export function userInfo(id: number) {
  return request.get<ApiResponse<UserInfoResp>>(`/users/${id}`)
}

export interface UserCreateReq {
  username: string
  email: string
  password: string
}

export function userCreate(body: UserCreateReq) {
  return request.post<ApiResponse>('/users', body)
}

export interface UserUpdateReq {
  username: string
  email: string
  password?: string
  status: boolean
}

export function userUpdate(id: number, body: UserUpdateReq) {
  return request.put<ApiResponse>(`/users/${id}`, body)
}

export interface UserUpdateStatusReq {
  status: boolean
}

export function userUpdateStatus(id: number, body: UserUpdateStatusReq) {
  return request.put<ApiResponse>(`/users/${id}/status`, body)
}

export function userDelete(id: number) {
  return request.delete<ApiResponse>(`/users/${id}`)
}
