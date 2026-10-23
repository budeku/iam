import request from '../index'
import type { ApiResponse } from './response'

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
