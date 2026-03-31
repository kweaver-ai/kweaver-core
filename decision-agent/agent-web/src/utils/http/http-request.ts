import axios from 'axios';
import type { Method } from 'axios';
import { curry } from 'lodash';
import qs from 'query-string';
import type { OptionsType, LangType } from './types';
import { config, businessDomainHeaderKey } from './types';
import { handleError } from './error-handler';
import axiosInstance from './axios-instance';

function convertLangType(lang: LangType): string {
  const [first, second] = lang.split('-');
  return [first.toLowerCase(), second.toUpperCase()].join('-');
}

function getConfig(key: string) {
  return (config as any)[key];
}

function setConfig(obj: Record<string, any>) {
  Object.keys(obj).forEach((key: string) => {
    (config as any)[key] = obj[key];
  });
}

/* 获取请求头部 */
function getCommonHttpHeaders() {
  const language = convertLangType(config.lang);
  return {
    Pragma: 'no-cache',
    Authorization: 'Bearer ' + config.getToken(),
    Token: config.getToken(),
    'Cache-Control': 'no-cache',
    'X-Requested-With': 'XMLHttpRequest',
    'x-language': language,
    'Accept-Language': language,
    [businessDomainHeaderKey]: config.businessDomainID,
  };
}

function getHttpBaseUrl() {
  const { protocol, host, port, prefix } = config;
  return `${protocol}//${host}:${port}${prefix}`;
}

const createHttpRequest = curry((method: string, url: string, options: OptionsType | undefined) => {
  const fullUrl = `${getHttpBaseUrl()}${url}`;
  const { body, headers, timeout = 60000, params, returnFullResponse } = options || {};

  const CancelToken = axios.CancelToken;
  let cancel: (message?: string) => void;

  const axiosConfig: any = {
    method: method.toLowerCase() as Method,
    url: fullUrl,
    data: body,
    params,
    paramsSerializer: {
      serialize: (params: any) => qs.stringify(params),
    },
    headers: {
      ...getCommonHttpHeaders(),
      ...headers,
    },
    timeout,
    cancelToken: new CancelToken(c => {
      cancel = c;
    }),
    // Axios 1.x 默认处理 JSON 转换，除非有特殊处理需求，否则无需自定义 transformRequest
    transformResponse: [
      (data: any) => {
        if (data && typeof data === 'string') {
          try {
            return JSON.parse(data);
          } catch {
            return data;
          }
        }
        return data;
      },
    ],
    validateStatus: (status: number) => status < 400,
  };

  const promise: any = new Promise((resolve, reject) => {
    (async () => {
      try {
        let response;

        switch (method.toLowerCase()) {
          case 'get':
          case 'post':
          case 'put':
          case 'patch':
          case 'delete':
            response = await axiosInstance.request(axiosConfig);
            break;
          default:
            throw new Error(`Unsupported HTTP method: ${method}`);
        }

        if (returnFullResponse) {
          // 有些场景，除了data还需要其它返回信息
          resolve(response);
        } else {
          resolve(response.data);
        }
      } catch (error) {
        handleError({
          error,
          url,
          reject,
          isOffline: !navigator.onLine,
        });
      }
    })();
  });

  promise.abort = () => cancel('CANCEL');
  return promise;
});

/**
 * 带有缓存功能的http
 */
const cacheableHttpFn = () => {
  const caches: Record<string, any> = {};

  return (method: string, url: string, options?: OptionsType & { expires?: number }) => {
    const { body, params, expires = -1 } = options || {};

    // 用 url + params + body 当做 key，来存储缓存
    const queryStr = qs.stringify(params || {});
    const key = `${method}:${url}${queryStr ? `?${queryStr}` : ''}${JSON.stringify(body || {})}`;

    if (!caches[key]) {
      caches[key] = createHttpRequest(method, url, options);

      if (expires !== -1) {
        setTimeout(() => {
          // 清除缓存
          delete caches[key];
        }, expires);
      }
    }

    return caches[key];
  };
};

const cacheableHttp = cacheableHttpFn();

export { createHttpRequest, cacheableHttp, setConfig, getConfig, getCommonHttpHeaders, getHttpBaseUrl };
