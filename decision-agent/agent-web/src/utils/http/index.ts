import {
  createHttpRequest,
  cacheableHttp,
  setConfig,
  getConfig,
  getCommonHttpHeaders,
  getHttpBaseUrl,
} from './http-request';
import { type OptionsType, businessDomainHeaderKey } from './types';
export * from './streaming-http';
export const get = (url: string, options?: OptionsType) => createHttpRequest('GET', url, options);
export const post = (url: string, options?: OptionsType) => createHttpRequest('POST', url, options);
export const put = (url: string, options?: OptionsType) => createHttpRequest('PUT', url, options);
export const del = (url: string, options?: OptionsType) => createHttpRequest('DELETE', url, options);
export const patch = (url: string, options?: OptionsType) => createHttpRequest('PATCH', url, options);
export const cacheableGet = (url: string, options?: OptionsType & { expires?: number }) =>
  cacheableHttp('GET', url, options);

export { setConfig, getConfig, getCommonHttpHeaders, getHttpBaseUrl, businessDomainHeaderKey };
