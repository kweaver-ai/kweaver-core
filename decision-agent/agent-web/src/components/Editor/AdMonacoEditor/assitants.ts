import DevTools from './monaco-themes/DevTools.json';
import Clouds from './monaco-themes/Clouds.json';
import Dolphin from './monaco-themes/dolphin.json';
import { loader } from '@monaco-editor/react';
import { getHttpBaseUrl, getConfig } from '@/utils/http';
import { registerDolphinLanguage } from './dolphin/registerDolphinLanguage';
import { Chrome_DevTools_Theme, Clouds_Theme, Dolphin_Theme } from '../static';

export const initAdMonacoEditor = () => {
  const baseUrl = getHttpBaseUrl();
  const lang = getConfig('lang');

  // 判断是否为独立运行环境
  const isStandalone = !(window as any).__POWERED_BY_QIANKUN__;
  // 独立运行时，资源路径直接从根路径开始；微前端环境下添加前缀
  const resourcePrefix = isStandalone ? '' : '/agent-web/public';

  // 使用本地资源，支持离线使用
  loader.config({
    paths: { vs: `${baseUrl}${resourcePrefix}/monaco/vs` },
    'vs/nls': {
      availableLanguages: { '*': lang === 'zh-cn' ? 'zh-cn' : lang === 'zh-tw' ? 'zh-tw' : '' }, // 国际化
    },
  });

  // 设置codicon字体(codicon字体文件的src，是在css中设置的，子应用不能在css中设置url，会导致路径不正确，所以需要在js中重新设置)
  loadAndApplyCodiconFont(baseUrl, resourcePrefix);

  // 定义主题, 定义完成的主题  可以在onMount的时候 使用
  loader.init().then(monaco => {
    monaco.editor.defineTheme(Clouds_Theme, Clouds);
    monaco.editor.defineTheme(Chrome_DevTools_Theme, DevTools);
    monaco.editor.defineTheme(Dolphin_Theme, Dolphin);
    // 注册ad自创的dolphin语言
    registerDolphinLanguage(monaco);
    // registerPythonLanguage(monaco);
  });
};

function loadAndApplyCodiconFont(baseUrl: string, resourcePrefix: string) {
  // 创建一个新的 FontFace 对象
  const font = new FontFace(
    'codicon',
    `url(${baseUrl}${resourcePrefix}/monaco/vs/base/browser/ui/codicons/codicon/codicon.ttf)`
  );

  // 加载字体
  font
    .load()
    .then(() => {
      // 字体加载成功后，将其添加到文档中
      document.fonts.add(font);
    })
    .catch(error => {
      // 处理字体加载失败的情况
      console.error('字体加载失败:', error);
    });
}
