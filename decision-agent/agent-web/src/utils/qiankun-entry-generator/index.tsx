import './public-path';
import '@/styles/main.less';
import React, { type ReactNode } from 'react';
import { createRoot } from 'react-dom/client';
import type { RouteObject } from 'react-router-dom';
import { ConfigProvider, message } from 'antd';
import { StyleProvider } from '@ant-design/cssinjs';
import zhTW from 'antd/es/locale/zh_TW';
import zhCN from 'antd/es/locale/zh_CN';
import enUS from 'antd/es/locale/en_US';
import { LangType } from '@/utils/http/types';
import { MicroWidgetContext } from '@/hooks/useMicroWidgetProps';
import { setConfig } from '@/utils/http';
import { initializeI18n } from '@/i18n';
import { initAdMonacoEditor } from '@/components/Editor/AdMonacoEditor/assitants';
import { RoutesComponent } from './RoutesComponent';
import '@/assets/graphIcons/iconfont.css';
import '@/assets/graphIconsMore/iconfont.css';
import { apis } from '@aishu-tech/components/dist/dip-components.min';
import '@aishu-tech/components/dist/dip-components.min.css';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import { mainAppProps } from './mockMainAppProps.ts';
dayjs.locale('zh-cn');

// 常量定义
const APP_PREFIX_CLS = 'agent-web';
const DEFAULT_ROOT_ID = APP_PREFIX_CLS;

// 公共配置类型定义
export interface QiankunEntryOptions {
  /** 应用根元素ID */
  rootId?: string;
  /** 自定义配置函数 */
  customConfig?: (props: any, container: any) => void;
}

// 获取UI语言配置
function getUILocale(lang: LangType): typeof enUS | typeof zhTW | typeof zhCN {
  const langs = {
    [LangType.us]: enUS,
    [LangType.tw]: zhTW,
    [LangType.zh]: zhCN,
  };

  return langs[lang] || zhCN;
}

// 公共配置设置函数
function setupAppConfig(props?: any, container?: any) {
  const protocol = props.config.systemInfo.location.protocol;
  const host = props.config.systemInfo.location.hostname;
  const port = props.config.systemInfo.location.port || 443;
  const prefix = props.prefix || '';
  const lang = props.language.getLanguage;
  const getToken = () => props.token.getToken.access_token;
  const refreshToken = props.token.refreshOauth2Token;
  const onTokenExpired = props.token.onTokenExpired;
  const theme = props.config.getTheme.normal;
  const businessDomainID = props.businessDomainID;

  // 初始化国际化
  initializeI18n(lang);

  // 设置http请求所需的信息
  setConfig({
    protocol,
    host,
    port,
    lang,
    prefix,
    getToken,
    refreshToken,
    onTokenExpired,
    toast: message,
    theme,
    businessDomainID,
  });

  // 初始化AdMonacoEditor
  initAdMonacoEditor();

  // 设置dip-components所需的信息
  apis.setup({
    protocol,
    host,
    port,
    lang,
    prefix,
    getToken,
    refreshToken,
    onTokenExpired,
    theme,
    popupContainer: container,
  });
}

// 创建Antd配置提供器
const AntdConfigProvider: React.FC<{
  children: ReactNode;
  props: any;
  container: any;
}> = ({ children, props, container }) => {
  const lang = props.language.getLanguage;
  const theme = props.config.getTheme.normal;

  ConfigProvider.config({
    prefixCls: APP_PREFIX_CLS,
  });

  // 指定弹出容器
  message.config({
    getContainer: () => container,
  });

  return (
    <StyleProvider hashPriority="high">
      <ConfigProvider
        button={{
          autoInsertSpace: false,
        }}
        prefixCls={APP_PREFIX_CLS}
        locale={getUILocale(lang)}
        theme={{
          token: {
            colorPrimary: theme,
          },
          components: {
            Tooltip: {
              colorBgSpotlight: 'rgb(72, 77, 101)', // 背景颜色
            },
          },
        }}
        getPopupContainer={() => container}
      >
        {children}
      </ConfigProvider>
    </StyleProvider>
  );
};

// 主生成器函数 - 创建普通应用
export function createEntry(AppComponent: React.ComponentType<any>, options: QiankunEntryOptions = {}) {
  const { rootId = DEFAULT_ROOT_ID, customConfig } = options;

  let root: any;

  // qiankun生命周期函数
  const bootstrap = async () => {};

  const mount = async (props: Record<string, any>) => {
    const container = props?.container?.querySelector(`#${rootId}`) || document.getElementById(rootId);

    // 执行公共配置
    setupAppConfig(props, container);

    // 执行自定义配置
    customConfig?.(props, container);

    // 创建根元素
    root = createRoot(container);

    // 渲染应用
    const AppWrapper = () => (
      <MicroWidgetContext.Provider value={{ ...props, container }}>
        <AntdConfigProvider props={props} container={container}>
          <AppComponent {...props} />
        </AntdConfigProvider>
      </MicroWidgetContext.Provider>
    );

    root.render(<AppWrapper />);
  };

  const unmount = async () => {
    try {
      root?.unmount();
    } catch (error) {
      console.error('Unmount error:', error);
    }
  };

  // 独立运行时自动挂载
  if (!(window as any).__POWERED_BY_QIANKUN__) {
    mount({
      ...mainAppProps,
      history: {
        ...mainAppProps.history,
        getBasePath: window.location.pathname.replace(/(\.html).*$/, '$1'),
      },
    });
  }

  return { bootstrap, mount, unmount };
}

// 路由应用生成器函数 - 创建路由应用
export function createRouteApp(routes: RouteObject[], options: QiankunEntryOptions = {}) {
  const App = (props: any) => {
    const routeBaseName = props.history?.getBasePath;
    return <RoutesComponent routes={routes} basename={routeBaseName} />;
  };

  return createEntry(App, options);
}
