import _ from 'lodash';
import { useEffect, useRef } from 'react';
import { useCreation } from './useCreation';
import { type Options } from './options';

type Fn = (...args: any) => any;

function useThrottleFn<T extends Fn>(fn: T, options?: Options) {
  const fnRef = useRef<T>(fn);
  fnRef.current = fn;

  const wait = options?.wait ?? 1000;

  const throttled = useCreation(
    () =>
      _.throttle<T>(
        ((...args: any[]) => {
          return fnRef.current(...args);
        }) as T,
        wait,
        options
      ),
    []
  );

  useEffect(() => {
    throttled.cancel();
  }, []);

  return {
    run: throttled as unknown as T,
    cancel: throttled.cancel,
    flush: throttled.flush
  };
}

export { useThrottleFn };
