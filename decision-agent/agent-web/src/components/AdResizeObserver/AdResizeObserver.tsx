import React, { PropsWithChildren, useEffect, useRef } from 'react';

export type ResizeProps = {
  width: number;
  height: number;
  dom: HTMLElement;
  visible: boolean; // 监听的元素在视口内是否处于可见状态
};

interface AdResizeObserverProps {
  onResize?: (data: ResizeProps) => void;
}

/**
 * 监听 children dom元素的尺寸变化
 * 注意：
 * children 必须只有一个父节点
 * @param children
 * @param onResize
 * @constructor
 */
const AdResizeObserver: React.FC<PropsWithChildren<AdResizeObserverProps>> = ({ children, onResize }) => {
  const resizeObserverRef = useRef<ResizeObserver>();
  const domRef = useRef<HTMLElement | null>(null);
  useEffect(() => {
    createResizeObserver();
    return () => {
      destroyResizeObserver();
    };
  }, []);

  const createResizeObserver = () => {
    resizeObserverRef.current = new ResizeObserver(entries => {
      if (domRef.current) {
        const { width, height } = domRef.current.getBoundingClientRect();
        onResize && onResize({ width, height, dom: domRef.current, visible: width !== 0 });
      }
    });
    resizeObserverRef.current?.observe(domRef.current as Element);
  };

  const destroyResizeObserver = () => {
    resizeObserverRef.current?.disconnect();
  };

  return React.cloneElement(children as React.ReactElement, {
    ref: domRef
  });
};

export default AdResizeObserver;
