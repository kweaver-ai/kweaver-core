/**
 * Axios 实例配置
 */
import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios';
import { getApiBaseUrl, HTTP_TIMEOUT } from '@/constants/api';

/** 创建 API 客户端实例 */
export const apiClient: AxiosInstance = axios.create({
  baseURL: getApiBaseUrl(),
  timeout: HTTP_TIMEOUT,
  headers: {
    'Content-Type': 'application/json',
  },
});

/** 请求拦截器 */
apiClient.interceptors.request.use(
  (config) => {
    // Dynamically set baseURL from runtime config
    config.baseURL = getApiBaseUrl();

    // Debug logging for API URL detection (can be disabled in production)
    if (import.meta.env.DEV) {
      console.debug(`[API] Request to: ${config.baseURL}${config.url}`);
    }

    // 可以在这里添加认证 token
    // const token = localStorage.getItem('token');
    // if (token) {
    //   config.headers.Authorization = `Bearer ${token}`;
    // }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

/** 响应拦截器 */
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    // 直接返回响应数据
    return response.data;
  },
  (error) => {
    // 统一错误处理
    if (error.response) {
      // 服务器返回错误状态码
      const { status, data } = error.response;
      console.error(`API Error [${status}]:`, data);
    } else if (error.request) {
      // 请求已发出但没有收到响应
      console.error('Network Error:', error.message);
    } else {
      // 请求配置错误
      console.error('Request Error:', error.message);
    }
    return Promise.reject(error);
  }
);

export default apiClient;
