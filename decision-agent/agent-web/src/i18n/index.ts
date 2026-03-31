import intl from 'react-intl-universal';
import locales from './locales';

function formatLocaleString(locale?: string) {
  if (!locale) return locale;

  // 将字符串按 '-' 分割成数组
  const parts = locale.split('-');

  // 如果格式不正确（没有连字符或不止一个连字符），返回原字符串
  if (parts.length !== 2) {
    return locale;
  }

  // 将第二部分转换为大写
  parts[1] = parts[1].toUpperCase();

  // 重新组合两部分
  return parts.join('-');
}

function initializeI18n(lng?: string) {
  const formatLang = formatLocaleString(lng);

  // 初始化语言
  intl.init({
    currentLocale: formatLang,
    locales,
    warningHandler: () => '',
  });
}

export { initializeI18n };
