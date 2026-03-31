import axios from 'axios';
import { config } from './types';

// 是否正在刷新token的标记
let isRefreshing = false;
let requests: Array<(token: string) => void> = [];

const axiosInstance = axios.create();

// 错误响应拦截（401 刷新token）
axiosInstance.interceptors.response.use(
  res => res,
  async err => {
    if (config.refreshToken && err.response && err.response.status === 401) {
      try {
        if (!isRefreshing) {
          isRefreshing = true;

          const token = await config.refreshToken();
          const newToken = token ? token.access_token : config.getToken();

          if (newToken) {
            requests.forEach(cb => cb(newToken));
            requests = [];

            return axiosInstance.request({
              ...err.config,
              headers: Object.assign({}, err.config.headers, {
                Authorization: `Bearer ${newToken}`,
              }),
            });
          }

          throw err;
        }

        return new Promise(resolve => {
          // 将resolve放进队列，用一个函数形式来保存，等token刷新后直接执行
          requests = [
            ...requests,
            token =>
              resolve(
                axiosInstance.request({
                  ...err.config,
                  headers: Object.assign({}, err.config.headers, {
                    Authorization: `Bearer ${token}`,
                  }),
                })
              ),
          ];
        });
      } catch {
        isRefreshing = false;
        throw err;
      } finally {
        if (!requests.length) {
          isRefreshing = false;
        }
      }
    } else {
      throw err;
    }
  }
);

export default axiosInstance;
