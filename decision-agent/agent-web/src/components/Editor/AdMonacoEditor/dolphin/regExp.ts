import { dolphinKeywords } from '../../static';
// 注释
export const commentRegExp = /^#.*$/;
// 变量
export const variableRegExp = /\$[a-zA-Z_][a-zA-Z0-9_]*/;

// 调用函数 @ 方式
export const callFuncRegExp = /@[^\s(]+\([^)]*\)/;

// 调用函数 tools= 方式
export const toolsFuncRegExp = /tools=\[(.*?)\]/;

// 关键字
const keywords = dolphinKeywords.map(keyword => keyword.insertText);
export const keywordRegExp = new RegExp(`\\b(${keywords.join('|')})\\b`);

// import语句
export const importRegExp = /^import.*$/;

// 字符串
export const stringRegExp = /(".*?"|'.*?')/;
// 数字
export const numberRegExp = /\d+(\.\d+)?/;
// 分隔符支持
export const delimiterRegExp = /[，。；：？！""'']/;
// 普通文本
export const plainTextRegExp = /[a-zA-Z_]+[a-zA-Z0-9_]*/;
