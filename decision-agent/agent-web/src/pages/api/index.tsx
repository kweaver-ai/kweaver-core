import { lazy } from 'react';
import { createRouteApp } from '@/utils/qiankun-entry-generator';
import { ModeEnum } from '@/components/DecisionAgent/types';

const routeComponents = {
  DecisionAgent: lazy(() => import('@/components/DecisionAgent')),
};

const routes = [
  {
    path: '/',
    element: <routeComponents.DecisionAgent mode={ModeEnum.API} />,
  },
];

const { bootstrap, mount, unmount } = createRouteApp(routes);
export { bootstrap, mount, unmount };
