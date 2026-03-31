import { Suspense, useMemo } from 'react';
import { Spin } from 'antd';
import { createBrowserRouter, RouterProvider, type RouteObject } from 'react-router-dom';

// 路由组件 - 专门用于qiankun-entry-generator
interface RoutesComponentProps {
  routes: RouteObject[];
  basename?: string;
}

export function RoutesComponent({ routes, basename }: RoutesComponentProps) {
  const router = useMemo(() => createBrowserRouter(routes, { basename: basename || '/' }), [routes, basename]);

  return (
    <Suspense
      fallback={
        <div className="dip-position-center">
          <Spin size="large" />
        </div>
      }
    >
      <RouterProvider router={router} />
    </Suspense>
  );
}
