import type { DependencyList, useMemo as useMemoType } from 'react';
import { useMemo, useRef } from 'react';
import isEqual from 'lodash/isEqual';
import cloneDeep from 'lodash/cloneDeep';

type MemoHookType = typeof useMemoType;
type DeepCompareMemoType = <T>(factory: () => T, deps?: DependencyList, deep?: boolean) => any;
type CreateUpdateMemo = (hook: MemoHookType) => DeepCompareMemoType;

/**
 * useMemo 的深度比较版本, 多出来第三个参数，deep，意思：是否采用深拷贝的方式缓存上一次的依赖，默认true
 * 对于依赖项含有 useRef 或者 window.XXX 这种数据时, deep 务必开启
 * 应用场景：当依赖性含有引用类型的数据时，可使用此hook去代替useMemo
 */
const createDeepCompareMemo: CreateUpdateMemo =
  hook =>
  (factory, deps, deep = true) => {
    const lastDepsRef = useRef<DependencyList>(); // 缓存上一次的依赖项
    const signalRef = useRef<number>(0); // 用于触发useMemo的基本数据类型

    if (deps === undefined || !isEqual(lastDepsRef.current, deps)) {
      lastDepsRef.current = deep ? cloneDeep(deps) : deps;
      signalRef.current += 1;
    }

    return hook(factory, [signalRef.current]);
  };

const useDeepCompareMemo = createDeepCompareMemo(useMemo);

export default useDeepCompareMemo;
