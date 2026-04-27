/**
 * 应用入口文件
 */
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './App';
import { registerDebugHelpers } from '@/utils/config';

const container = document.getElementById('root');

if (container) {
  // Register debug helpers for production troubleshooting
  registerDebugHelpers();

  const root = createRoot(container);
  root.render(
    <StrictMode>
      <App />
    </StrictMode>
  );
}
