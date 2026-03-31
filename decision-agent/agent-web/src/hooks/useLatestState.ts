import type { Dispatch, SetStateAction } from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';

export type GetStateAction<S> = () => S;
export type ResetStateAction = (key?: string | string[]) => void;
type LatestStateFunc = <S>(initialState: S) => [S, Dispatch<SetStateAction<S>>, GetStateAction<S>, ResetStateAction];

/**
 * 与 useState 用法一致，多出了一个 getLatestState 方法, 同时也会保证在组件卸载的时候不会进行setState的执行
 * getLatestState方法可以获取到最新的state
 * 使用场景：
 * 解决 react hook 过时闭包导致拿不到 useState 中最新的 state 问题，
 * 例如：使用第三方库绑定的事件中，用到了state，当state更新后，绑定的第三方库的事件中获取不到更新后的state
 * @param initialState
 */
const useLatestState: LatestStateFunc = initialState => {
  const [value, setValue] = useState(initialState);
  const latestStateRef = useRef(value); // 缓存最新的state
  const unMountedRef = useRef<boolean>(false); // 记录有没有被卸载

  useEffect(() => {
    unMountedRef.current = false;
    return () => {
      unMountedRef.current = true;
    };
  }, []);

  const getState = useCallback(() => latestStateRef.current, []);

  const setState = useCallback((currentState: any) => {
    if (!unMountedRef.current) {
      const stateData = typeof currentState === 'function' ? currentState(latestStateRef.current) : currentState;
      latestStateRef.current = stateData;
      setValue(stateData);
    }
  }, []);
  const resetState = useCallback((key?: string | string[]) => {
    let stateData: any = {};
    if (key) {
      const keys = typeof key === 'string' ? [key] : key;
      keys.forEach(item => {
        stateData[item] = (initialState as Record<string, unknown>)[item];
      });
    } else {
      stateData = initialState;
    }
    setState((preState: any) => ({
      ...preState,
      ...stateData,
    }));
  }, []);
  return [value, setState, getState, resetState];
};
export default useLatestState;
