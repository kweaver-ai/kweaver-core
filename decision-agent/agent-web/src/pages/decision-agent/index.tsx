import { lazy } from 'react';
import { createRouteApp } from '@/utils/qiankun-entry-generator';

const routeComponents = {
  DecisionAgent: lazy(() => import('@/components/DecisionAgent')),
  AgentConfig: lazy(() => import('@/components/AgentConfig')),
  AgentUsage: lazy(() => import('./AgentUsage')),
  DolphinLanguageDoc: lazy(() => import('./DolphinLanguageDoc')),
  AgentApiDocument: lazy(() => import('./AgentApiDocument')),
};

const routes = [
  {
    path: '/',
    element: <routeComponents.DecisionAgent />,
  },
  {
    path: '/config',
    element: <routeComponents.AgentConfig />,
  },
  {
    path: '/usage',
    element: <routeComponents.AgentUsage />,
  },
  {
    path: '/dolphin-language-doc',
    element: <routeComponents.DolphinLanguageDoc />,
  },
  {
    path: '/api-doc',
    element: <routeComponents.AgentApiDocument />,
  },
];

const { bootstrap, mount, unmount } = createRouteApp(routes);
export { bootstrap, mount, unmount };
