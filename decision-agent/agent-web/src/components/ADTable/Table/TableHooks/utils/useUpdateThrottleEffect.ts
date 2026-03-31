import { createUpdateEffect } from './createUpdateEffect';
import { useThrottleEffect } from './useThrottleEffect';

const useUpdateThrottleEffect = createUpdateEffect(useThrottleEffect);

export { useUpdateThrottleEffect };
