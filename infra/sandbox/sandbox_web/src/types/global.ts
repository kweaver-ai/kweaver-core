/**
 * 全局类型定义
 */

/** 路由路径 */
export type RoutePath = '/' | '/templates' | '/sessions' | '/execute' | '/files';

/** 应用菜单项 */
export interface MenuItem {
  key: string;
  label: string;
  path: RoutePath;
  icon: string;
}

/** 应用状态 */
export interface AppState {
  currentRoute: RoutePath;
  sidebarCollapsed: boolean;
}
