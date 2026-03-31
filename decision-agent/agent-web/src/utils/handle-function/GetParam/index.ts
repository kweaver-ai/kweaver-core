/**
 * @description 获取url中的参数
 * @param {string} name 参数名
 */

function getParam(name?: string | string[]): any {
  if (!window) return '';

  let paramsObj: Record<string, string> = {};

  if (URLSearchParams) {
    paramsObj = Object.fromEntries(new URLSearchParams(window.location.search).entries());
  } else {
    const reg = /[?&]([^=&#]+)=([^=&#]*)/g;
    window.location.search.replace(reg, (_, $1, $2) => {
      paramsObj[decodeURIComponent($1)] = decodeURIComponent($2);
      return _;
    });
  }

  if (!name) return paramsObj;
  if (typeof name === 'string') return paramsObj[name] || '';
  if (Array.isArray(name)) {
    return name.reduce((res, key) => ({ ...res, [key]: paramsObj[key] || '' }), {});
  }

  return paramsObj;
}

/**
 * 获取search参数中的参数
 * @param url search参数
 * @returns 所有参数组成的对象
 */
const getUrl = (url: any) => {
  const returl: any = {};
  if (url.indexOf('?') !== -1) {
    const queryString = url.substr(1);
    const queryParams = queryString.split('&');
    for (let i = 0; i < queryParams.length; i++) {
      const [key, value] = queryParams[i].split('=');
      returl[key] = decodeURI(value);
    }
  }
  return returl;
};

export { getParam, getUrl };
