import { useRef, useEffect } from 'react';
import { useBlocker } from 'react-router-dom';

interface useNavigationBlockerProps {
  shouldBlock: boolean;
  handleNavigation: (blocker: ReturnType<typeof useBlocker>) => void;
}

const useNavigationBlocker = ({ shouldBlock, handleNavigation }: useNavigationBlockerProps) => {
  const shouldBlockRef = useRef(shouldBlock);

  // 使用 ref 来保持最新的 shouldBlock 值
  useEffect(() => {
    shouldBlockRef.current = shouldBlock;
  }, [shouldBlock]);

  // 处理 React Router 路由跳转拦截
  const blocker = useBlocker(({ currentLocation, nextLocation }) => {
    // 只有当有未保存更改且确实是导航到不同路由时才拦截
    return shouldBlockRef.current && currentLocation.pathname !== nextLocation.pathname;
  });

  useEffect(() => {
    if (blocker.state === 'blocked' && shouldBlockRef.current) {
      handleNavigation(blocker);
    }
  }, [blocker, handleNavigation]);

  // 处理浏览器刷新和关闭
  useEffect(() => {
    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      if (shouldBlockRef.current) {
        event.preventDefault();
        // 现代浏览器中，这个字符串可能不会显示，但仍需要设置
        event.returnValue = '您有未保存的更改，确定要离开吗？';
        return '您有未保存的更改，确定要离开吗？';
      }
    };

    // 监听页面刷新和关闭
    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, []);
};

export default useNavigationBlocker;
