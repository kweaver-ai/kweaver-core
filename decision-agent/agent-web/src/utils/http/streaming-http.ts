import { EventSourceMessage, fetchEventSource } from '@microsoft/fetch-event-source';
import { isJSONString } from '@/utils/handle-function';
import _ from 'lodash';
import { config, LangType, IncrementalActionEnum, businessDomainHeaderKey } from './types';
export type StreamingOutServerType = {
  url: string; // 请求的URL
  method?: 'POST' | 'GET'; // 请求方式 默认post方式
  body: any; // 请求的参数
  onMessage?: (event: EventSourceMessage) => void; // 接收一次数据段时回调，因为是流式返回，所以这个回调会被调用多次
  onClose?: () => void; // 正常结束的回调
  onError?: (error: any) => void; // 各种错误最终都会走这个回调
  onOpen?: (controller: AbortController) => void; // 建立连接的时候
};

function convertLangType(lang: LangType): string {
  const [first, second] = lang.split('-');
  return [first.toLowerCase(), second.toUpperCase()].join('-');
}

function getHttpBaseUrl() {
  const { protocol, host, port, prefix } = config;
  return `${protocol}//${host}:${port}${prefix}`;
}

const getStreamingOutHttpHeaders = () => {
  return {
    Authorization: 'Bearer ' + config.getToken(),
    'Content-Type': 'application/json; charset=utf-8',
    Connection: 'keep-alive',
    responseType: 'text/event-stream',
    'x-language': convertLangType(config.lang),
    [businessDomainHeaderKey]: config.businessDomainID,
  };
};

/** 流式http请求 （所有流式请求统一使用这个发起） */
export const streamingOutHttp = (param: StreamingOutServerType): AbortController => {
  const { url, body, method = 'POST', onMessage, onError, onClose, onOpen } = param;
  const controller = new AbortController();
  const signal = controller.signal;
  let errorInfo = {};
  const fullUrl = `${getHttpBaseUrl()}${url}`;
  fetchEventSource(fullUrl, {
    signal,
    headers: getStreamingOutHttpHeaders(),
    body: JSON.stringify(body),
    method,
    openWhenHidden: true,
    // 建立连接的回调
    async onopen(response: Response) {
      if (!response.ok) {
        if (response.status === 401) {
          // 说明token过期， 要自动续
          await config.refreshToken?.();
          streamingOutHttp(param);
          return;
        }
        // 说明建立连接异常
        const reader = response.body?.getReader();
        const textDecoder = new TextDecoder('utf-8');
        const chunk = await reader?.read();
        const valueError = textDecoder.decode(chunk?.value);
        console.log(valueError, '流式接口报错了');
        const description =
          typeof valueError === 'string'
            ? isJSONString(valueError)
              ? JSON.parse(valueError)
              : valueError
            : valueError;
        errorInfo = { error: description, code: response.status };
        console.log(errorInfo, 'errorInfo');
        throw new Error(valueError);
      }
      onOpen?.(controller);
    },
    // 接收一次数据段时回调，因为是流式返回，所以这个回调会被调用多次
    onmessage: (event: EventSourceMessage) => {
      onMessage?.(event);
    },
    // 正常结束的回调
    onclose: () => {
      controller.abort();
      onClose?.();
    },
    // 连接出现异常回调
    onerror: (error: Error) => {
      controller.abort();
      onError?.(errorInfo);
      if (!_.isEmpty(error)) {
        throw new Error(error.message);
      } else {
        throw new Error('报错了');
      }
    },
  });
  return controller;
};

/**
 * 处理增量流式数据更新
 * @param {Object} params - 更新参数
 * @param {number} params.seq_id - 序列ID
 * @param {string[]} params.key - 数据路径键
 * @param {any} params.content - 更新内容
 * @param {'upsert'|'append'|'remove'|'end'} params.action - 操作类型
 * @param {Object|string} originalData - 原始数据
 * @returns {Object|string} 更新后的数据
 */
export function processIncrementalUpdate(
  {
    key: pathKeys,
    content: newContent,
    action: operation,
  }: {
    seq_id: number;
    key: string[];
    content: any;
    action: IncrementalActionEnum;
  },
  originalData: object | string
): object | string {
  // 处理根路径操作
  if (pathKeys.length === 0) {
    switch (operation) {
      case IncrementalActionEnum.Upsert:
        return newContent;
      case IncrementalActionEnum.Append:
        return typeof originalData === 'string' ? originalData + newContent : originalData;
      default:
        return originalData;
    }
  }

  // 处理嵌套路径操作
  switch (operation) {
    case IncrementalActionEnum.Upsert:
      _.set(originalData as object, pathKeys, newContent);
      return originalData;

    case IncrementalActionEnum.Append: {
      const existingValue = _.get(originalData, pathKeys);
      const updatedValue = typeof existingValue === 'string' ? existingValue + newContent : newContent;
      _.set(originalData as object, pathKeys, updatedValue);
      return originalData;
    }

    case IncrementalActionEnum.Remove: {
      const parentPath = pathKeys.slice(0, -1);
      const lastKey = pathKeys[pathKeys.length - 1];
      const parent = parentPath.length > 0 ? _.get(originalData, parentPath) : originalData;

      if (Array.isArray(parent)) {
        // 刪除元素 要保留元素在數組中的位置
        delete parent[Number(lastKey)];
        // parent.splice(parseInt(lastKey), 1, undefined);
      } else {
        _.unset(originalData, pathKeys);
      }
      return originalData;
    }

    default:
      return originalData;
  }
}
