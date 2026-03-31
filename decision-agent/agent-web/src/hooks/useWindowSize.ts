import { useEffect, useState } from 'react';
import _ from 'lodash';

export default function useWindowSize(): { width: number; height: number } {
  const [size, setSize] = useState({ width: window?.innerWidth || 0, height: window?.innerHeight || 0 });

  useEffect(() => {
    if (!window) return;

    const handleResize = _.throttle(() => {
      setSize({
        width: window.innerWidth,
        height: window.innerHeight,
      });
    }, 300);

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return size;
}
