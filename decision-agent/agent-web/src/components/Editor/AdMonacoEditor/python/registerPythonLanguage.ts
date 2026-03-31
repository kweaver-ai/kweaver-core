import { Monaco } from '@monaco-editor/react';
import { pythonLanguageConfig } from './enum';

export const registerPythonLanguage = (monaco: Monaco) => {
  // @ts-ignore - Simplified implementation for completion provider
  monaco.languages.registerCompletionItemProvider('python', {
    provideCompletionItems: (model: any, position: any) => {
      const suggestions = [
        ...pythonLanguageConfig.builtins.map(func => ({
          label: func,
          kind: 2,
          insertText: func,
          detail: '内置函数'
        })),
        ...pythonLanguageConfig.keywords.map(keyword => ({
          label: keyword,
          kind: 14,
          insertText: keyword,
          detail: '关键字'
        })),
        // 常用代码片段
        {
          label: 'def',
          kind: 27,
          // eslint-disable-next-line no-template-curly-in-string
          insertText: ['def ${1:function_name}(${2:parameters}):', '\t${3:pass}'].join('\n'),
          insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          detail: '函数定义',
          documentation: '创建一个新函数'
        },
        {
          label: 'if',
          kind: 27,
          // eslint-disable-next-line no-template-curly-in-string
          insertText: ['if ${1:condition}:', '\t${2:pass}'].join('\n'),
          insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          detail: '条件语句',
          documentation: '创建一个if条件语句'
        }
      ];
      return { suggestions };
    }
  });

  // 注册Python语言的语法高亮规则
  // @ts-ignore - Simplified monarch language provider
  monaco.languages.setMonarchTokensProvider('python', pythonLanguageConfig);
};
