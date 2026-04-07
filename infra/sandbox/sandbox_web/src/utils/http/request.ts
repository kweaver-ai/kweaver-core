/**
 * HTTP 请求封装工具
 */
import apiClient from './axios';
import type { AxiosRequestConfig } from 'axios';

/**
 * 通用 GET 请求
 */
export function get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return apiClient.get<T>(url, config);
}

/**
 * 通用 POST 请求
 */
export function post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return apiClient.post<T>(url, data, config);
}

/**
 * 通用 PUT 请求
 */
export function put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return apiClient.put<T>(url, data, config);
}

/**
 * 通用 DELETE 请求
 */
export function del<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return apiClient.delete<T>(url, config);
}

/**
 * 通用 PATCH 请求
 */
export function patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return apiClient.patch<T>(url, data, config);
}
