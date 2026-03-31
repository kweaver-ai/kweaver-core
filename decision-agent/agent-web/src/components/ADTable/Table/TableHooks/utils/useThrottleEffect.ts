import { type DependencyList, type EffectCallback, useEffect, useState } from 'react';
import { useThrottleFn } from './useThrottleFn';
import { useUpdateEffect } from './useUpdateEffect';
import { type Options } from './options';

function useThrottleEffect(effect: EffectCallback, deps?: DependencyList, options?: Options) {
  const [flag, setFlag] = useState({});

  const { run, cancel } = useThrottleFn(() => {
    setFlag({});
  }, options);

  useEffect(() => {
    return run();
  }, deps);

  useEffect(cancel, []);

  useUpdateEffect(effect, [flag]);
}

export { useThrottleEffect };
