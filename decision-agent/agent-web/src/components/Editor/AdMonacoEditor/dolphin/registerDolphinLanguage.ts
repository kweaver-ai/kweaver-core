import { Monaco } from '@monaco-editor/react';
import {
  callFuncRegExp,
  commentRegExp,
  delimiterRegExp,
  importRegExp,
  keywordRegExp,
  numberRegExp,
  plainTextRegExp,
  stringRegExp,
  toolsFuncRegExp,
  variableRegExp
} from './regExp';

export const dolphinAutoClosingPairs = [
  { open: '{', close: '}' },
  { open: '[', close: ']' },
  { open: '(', close: ')' },
  { open: '"', close: '"' },
  { open: "'", close: "'" }
];

export const registerDolphinLanguage = (monaco: Monaco) => {
  const languageName = 'dolphin';

  monaco.languages.register({ id: languageName });

  // 定义词法分析器
  monaco.languages.setMonarchTokensProvider(languageName, {
    operators: ['=', '>', '<', '==', '<=', '>=', '!=', '<>', '+', '-', '*', '/', '->', '>>'],
    tokenizer: {
      root: [
        [commentRegExp, 'comment'], // 单行注释
        [variableRegExp, 'variable'], // 变量
        [keywordRegExp, 'keyword'], // 关键字
        [stringRegExp, 'string'], // 字符串
        [numberRegExp, 'number'], // 数字
        [delimiterRegExp, 'delimiter'], // 分隔符
        // [importRegExp, 'importTool'], // 导入工具语句
        [callFuncRegExp, 'callFunc'], // 调用工具@方式
        [toolsFuncRegExp, 'toolsFunc'], // 调用工具tools=方式
        [plainTextRegExp, 'plainText'] // 普通文本
      ]
    },
    ignoreCase: true
  });

  // 定义语言的配置
  monaco.languages.setLanguageConfiguration(languageName, {
    comments: {
      lineComment: '#'
    },
    brackets: [
      ['{', '}'],
      ['[', ']'],
      ['(', ')']
    ],
    autoClosingPairs: dolphinAutoClosingPairs
  });
};
