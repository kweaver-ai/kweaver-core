import { isString } from '../isType';

/**
 * 基于indexOf的模糊匹配, 如果用正则, 对 * + ? 等特殊字符比较难处理
 * @param {String} keyWord 关键字
 * @param {String} text 文本
 */
const fuzzyMatch = (keyWord, text) => {
  if (!isString(keyWord) && !isString(text)) return;

  const k = keyWord?.toLowerCase();
  const t = text?.toLowerCase();

  return t.indexOf(k) !== -1;
};

export { fuzzyMatch };
