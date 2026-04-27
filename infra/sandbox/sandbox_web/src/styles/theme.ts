/**
 * Ant Design 主题配置
 * 基于 Figma 设计规范
 */
import type { ThemeConfig } from 'antd';

/** Sandbox 主题配置 */
export const sandboxTheme: ThemeConfig = {
  token: {
    // 主色
    colorPrimary: '#126ee3',
    colorPrimaryHover: '#0f5dc2',
    colorPrimaryActive: '#0b4fa8',

    // 背景色
    colorBgLayout: '#fafafa',
    colorBgContainer: '#ffffff',
    colorBgElevated: '#ffffff',

    // 边框色
    colorBorder: '#e7edf7',
    colorBorderSecondary: '#d9d9d9',

    // 文字颜色
    colorText: 'rgba(0,0,0,0.85)',
    colorTextSecondary: 'rgba(0,0,0,0.65)',
    colorTextTertiary: '#677489',
    colorTextQuaternary: 'rgba(0,0,0,0.45)',
    colorTextPlaceholder: 'rgba(0,0,0,0.25)',

    // 功能色
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: '#1890ff',

    // 圆角
    borderRadius: 4,
    borderRadiusLG: 8,
    borderRadiusSM: 4,
    borderRadiusXS: 2,

    // 字体大小
    fontSize: 14,
    fontSizeHeading1: 20,
    fontSizeHeading2: 16,
    fontSizeHeading3: 15,
    fontSizeHeading4: 14,
    fontSizeLG: 16,
    fontSizeSM: 12,
    fontSizeXL: 24,

    // 行高
    lineHeight: 1.5715,
    lineHeightLG: 1.5,
    lineHeightSM: 1.66,

    // 间距
    marginXS: 8,
    marginSM: 12,
    margin: 16,
    marginMD: 20,
    marginLG: 24,
    marginXL: 32,

    // 其他
    wireframe: false,
  },
  components: {
    Layout: {
      headerBg: '#ffffff',
      headerHeight: 64,
      headerPadding: '0 24px',
      headerColor: 'rgba(0,0,0,0.85)',
      siderBg: '#ffffff',
      bodyBg: '#fafafa',
    },

    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: 'rgba(18,110,227,0.06)',
      itemSelectedColor: '#126ee3',
      itemHoverBg: 'rgba(0,0,0,0.02)',
      itemActiveBg: 'rgba(18,110,227,0.1)',
      itemPaddingInline: 16,
      itemMarginBlock: 0,
      itemHeight: 40,
      itemBorderRadius: 4,
    },

    Table: {
      headerBg: '#fafafa',
      headerSplitColor: '#e7edf7',
      borderColor: '#e7edf7',
      headerColor: 'rgba(0,0,0,0.85)',
      cellPaddingInline: 16,
      cellPaddingBlock: 12,
      fontSize: 14,
      borderWidth: 1,
    },

    Button: {
      primaryShadow: 'none',
      defaultShadow: 'none',
      dashedShadow: 'none',
      linkShadow: 'none',
      textShadow: 'none',
      primaryColor: '#ffffff',
      defaultColor: 'rgba(0,0,0,0.85)',
      defaultBg: '#ffffff',
      defaultBorderColor: '#d9d9d9',
      defaultHoverBorderColor: '#126ee3',
      defaultHoverColor: '#126ee3',
      controlHeight: 36,
      controlHeightLG: 40,
      controlHeightSM: 24,
      paddingInline: 16,
      fontWeight: 400,
    },

    Input: {
      colorBorder: '#d9d9d9',
      colorBorderHover: '#126ee3',
      activeBorderColor: '#126ee3',
      inputBorderColor: '#d9d9d9',
      paddingInline: 12,
      controlHeight: 36,
      controlHeightLG: 40,
      controlHeightSM: 24,
      colorTextPlaceholder: 'rgba(0,0,0,0.25)',
      borderRadius: 4,
    },

    Select: {
      colorBorder: '#d9d9d9',
      colorBorderHover: '#126ee3',
      activeBorderColor: '#126ee3',
      optionSelectedBg: 'rgba(18,110,227,0.06)',
      controlHeight: 36,
      controlHeightLG: 40,
      controlHeightSM: 24,
      borderRadius: 4,
    },

    Modal: {
      contentBg: '#ffffff',
      headerBg: '#ffffff',
      headerBorderSplit: 0,
      footerBg: 'transparent',
      footerBorderSplit: 0,
      borderRadiusLG: 12,
    },

    Card: {
      colorBorderSecondary: '#e7edf7',
      borderRadiusLG: 12,
      paddingLG: 24,
      paddingMD: 16,
    },

    Tag: {
      borderRadiusSM: 4,
      defaultBg: '#fafafa',
      defaultColor: 'rgba(0,0,0,0.65)',
    },

    Form: {
      itemMarginBottom: 24,
      verticalLabelPadding: '0 0 8px',
      labelFontSize: 14,
      labelColor: 'rgba(0,0,0,0.85)',
    },

    Tabs: {
      itemActiveColor: '#126ee3',
      itemSelectedColor: '#126ee3',
      inkBarColor: '#126ee3',
      itemHoverColor: '#126ee3',
    },
  },
};
