/**
 * Router 组件
 */
import { Suspense, lazy } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, Spin, App as AntdApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import LayoutComponent from '@components/Layout';
import { sandboxTheme } from '@/styles/theme';

// 懒加载页面组件
const SessionsPage = lazy(() => import('@pages/sessions'));
const TemplatesPage = lazy(() => import('@pages/templates'));
const ExecutePage = lazy(() => import('@pages/execute'));

/** 加载中组件 */
function PageLoading() {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
      }}
    >
      <Spin size="large" />
    </div>
  );
}

/** Router 组件 */
export function Router() {
  return (
    <BrowserRouter>
      <ConfigProvider theme={sandboxTheme} locale={zhCN}>
        <AntdApp>
          <Suspense fallback={<PageLoading />}>
            <Routes>
              <Route path="/" element={<LayoutComponent />}>
                <Route index element={<Navigate to="/sessions" replace />} />
                <Route path="sessions" element={<SessionsPage />} />
                <Route path="templates" element={<TemplatesPage />} />
                <Route path="execute" element={<ExecutePage />} />
              </Route>
            </Routes>
          </Suspense>
        </AntdApp>
      </ConfigProvider>
    </BrowserRouter>
  );
}
