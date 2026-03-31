import { useEffect } from 'react';
import { createUpdateEffect } from './createUpdateEffect';

const useUpdateEffect = createUpdateEffect(useEffect);
export { useUpdateEffect };
