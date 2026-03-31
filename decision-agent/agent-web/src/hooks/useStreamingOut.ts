import { useEffect, useRef, useState } from 'react';
import useLatestState from './useLatestState';
import useDeepCompareMemo from './useDeepCompareMemo';
import { processIncrementalUpdate, streamingOutHttp, type StreamingOutServerType } from '@/utils/http/streaming-http';
import { isJSONString } from '@/utils/handle-function';

export interface UseTypeOutConfig extends Omit<StreamingOutServerType, 'body' | 'url'> {
  timeout?: number; // 字符之间的延迟输出间隔，默认 10ms
  errorCode?: string | string[]; // 流式输出过程中的错误代码
  url?: string;
}

export type UseTypeOutResponse = {
  generating: boolean; // 是否正在生成
  pending: boolean; // 请求是否在等待响应中
  content: string; // 每次流式输出的内容
  error?: string; // 接口自身异常，例如 504
  cancel: boolean; // 是否主动取消流式输出
};

type StartParamType = {
  body: any; // 流式请求体
  url?: string; // 流式请求体url
  increase_stream?: boolean; // 是不是真流式
};

type UseTypeOutStartFunc = ({ body, url, increase_stream }: StartParamType) => void;
type UseTypeOutStopFunc = () => void;

type UseTypeOutState = [UseTypeOutResponse, UseTypeOutStartFunc, UseTypeOutStopFunc];

type TypeOutFunc = (initialState?: UseTypeOutConfig) => UseTypeOutState;

/**
 * 对话框流式输出文本hook
 */
const useStreamingOut: TypeOutFunc = config => {
  const { url, onOpen, onError, onClose, onMessage } = config || {};
  const [content, setContent] = useState<any>(null);
  const [error, setError, getError] = useLatestState<any>(null);
  const [streamStatus, setStreamStatus] = useLatestState({
    streamReqEnd: true, // 流式请求是否结束
    generating: false, // 是否正在生成
    pending: false, // 请求是否在等待服务端响应中
    cancel: false, // 是否主动取消流式输出
  });
  const controllerRef = useRef<AbortController>();
  useEffect(() => {
    return () => {
      stop();
    };
  }, []);

  const handleMessage = (e: any) => {
    try {
      if (isJSONString(e.data)) {
        setContent(e.data);
      }
    } catch (error) {
      console.log(error);
    }
  };

  const start = (param: StartParamType) => {
    // 每次新发起请求之前，先取消之前的请求
    controllerRef.current?.abort();

    const convertContent: any = {};
    setContent(null);
    setError(null);

    setStreamStatus(prevState => ({
      ...prevState,
      streamReqEnd: false,
      generating: true,
      pending: true,
      cancel: false,
    }));
    const streamUrl = param.url ?? url;
    controllerRef.current = streamingOutHttp({
      url: streamUrl as string,
      body: param.body,
      onOpen: controller => {
        onOpen?.(controller);
      },
      onMessage: event => {
        setStreamStatus(prevState => ({
          ...prevState,
          pending: false,
        }));
        // 处理增量真流式
        if (param.increase_stream) {
          const eventData = JSON.parse(event.data);
          if (eventData?.BaseError?.error_code) {
            console.log('useStreamingOut onMessage Error', eventData);
            setError(eventData);
          } else {
            processIncrementalUpdate(eventData, convertContent);
            setContent(JSON.stringify(convertContent));
          }
        } else {
          handleMessage(event);
        }

        onMessage?.(event);
      },
      onClose: () => {
        setStreamStatus(prevState => ({
          ...prevState,
          streamReqEnd: true,
          generating: false,
        }));
        onClose?.();
      },
      onError: (errorInfo: any) => {
        console.log('报错了 useStreamingOut onError', errorInfo);
        if (!getError()) {
          setError(errorInfo?.error ?? '接口报错了');
        }
        setStreamStatus(prevState => ({
          ...prevState,
          streamReqEnd: true,
          generating: false,
          pending: false,
        }));
        onError?.(getError() || errorInfo);
      },
    });
  };

  const stop = () => {
    controllerRef.current?.abort();
    setStreamStatus(prevState => ({
      ...prevState,
      streamReqEnd: true,
      generating: false,
      pending: false,
      cancel: true,
    }));
  };

  const response: UseTypeOutResponse = useDeepCompareMemo(() => {
    return {
      content: content,
      generating: streamStatus.generating,
      pending: streamStatus.pending,
      cancel: streamStatus.cancel,
      error,
    };
  }, [streamStatus, content, error]);
  return [response, start, stop];
};
export default useStreamingOut;
