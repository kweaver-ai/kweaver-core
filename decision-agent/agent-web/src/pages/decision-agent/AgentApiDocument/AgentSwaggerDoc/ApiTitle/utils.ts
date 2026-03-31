/**
 * 插入、更新链接时添加http协议
 * @param url
 */
const customParseLinkUrl = (url: string): string => {
  if (url.indexOf('http') !== 0) {
    return `https://${url}`;
  }
  return url;
};

export const defaultEditorConfig = {
  MENU_CONF: {
    insertLink: {
      parseLinkUrl: customParseLinkUrl,
    },
    editLink: {
      parseLinkUrl: customParseLinkUrl,
    },
  },
};
