import screenfull from 'screenfull';

import katex from 'katex';
import 'katex/dist/katex.min.css';

import Cropper from 'cropperjs';
import 'cropperjs/dist/cropper.css';

import mermaid from 'mermaid';

import highlight from 'highlight.js';
import 'highlight.js/styles/atom-one-dark.css';

import * as prettier from 'prettier';
import parserMarkdown from 'prettier/plugins/markdown';
import { config } from 'md-editor-rt';
import LinkAttr from 'markdown-it-link-attributes';

export const initMarkdownConfig = () => {
  config({
    editorExtensions: {
      prettier: {
        prettierInstance: prettier,
        parserMarkdownInstance: parserMarkdown,
      },
      highlight: {
        instance: highlight,
      },
      screenfull: {
        instance: screenfull,
      },
      katex: {
        instance: katex,
      },
      cropper: {
        instance: Cropper,
      },
      mermaid: {
        instance: mermaid,
      },
    },
    markdownItPlugins(plugins) {
      return [
        ...plugins,
        {
          type: 'linkAttr',
          plugin: LinkAttr,
          options: {
            matcher(href: string) {
              // 如果使用了markdown-it-anchor
              // 应该忽略标题头部的锚点链接
              return !href.startsWith('#');
            },
            attrs: {
              target: '_blank',
            },
          },
        },
      ];
    },
  });
};

// 常见的编程语言和标记语言标识符列表
const validLanguages = [
  // 编程语言
  'javascript',
  'js',
  'typescript',
  'ts',
  'python',
  'py',
  'java',
  'c',
  'cpp',
  'c++',
  'go',
  'rust',
  'php',
  'ruby',
  'swift',
  'kotlin',
  'dart',
  'scala',
  'clojure',
  'haskell',
  'perl',
  'lua',
  'r',
  'matlab',
  'objective-c',
  'objc',
  'assembly',
  'asm',
  'vb',
  'vbnet',
  'fsharp',
  'fs',
  'elixir',
  'erlang',
  'nim',
  'crystal',
  'zig',
  'solidity',
  'move',
  'cairo',

  // 标记语言和配置文件
  'html',
  'css',
  'scss',
  'sass',
  'less',
  'json',
  'xml',
  'yaml',
  'yml',
  // 'markdown',
  'toml',
  'ini',
  'csv',

  // 数据库和查询语言
  'sql',
  'mysql',
  'postgresql',
  'sqlite',
  'mongodb',
  'graphql',

  // Shell 和脚本
  'bash',
  'sh',
  'zsh',
  'fish',
  'powershell',
  'ps1',
  'cmd',
  'bat',

  // 框架和库特定
  'vue',
  'jsx',
  'tsx',
  'svelte',
  'angular',

  // 配置和部署
  'dockerfile',
  'docker',
  'nginx',
  'apache',
  'makefile',
  'cmake',
  'jenkins',
  'github-actions',
  'gitlab-ci',

  // 其他
  // 'text',
  'plain',
  'diff',
  'patch',
  'log',
  'console',
  'terminal',
  'latex',
  'tex',
  'bibtex',
  'mermaid',
  'plantuml',
];

/**
 * 移除不符合要求的 Markdown 代码块语法标记
 * @param markdownStr Markdown 字符串
 * @returns 处理后的字符串
 */
function removeInvalidCode(markdownStr: string): string {
  // 将语言列表转换为Set，便于快速查找
  const validLanguageSet = new Set(validLanguages);

  // 按行分割字符串
  const lines = markdownStr.split('\n');
  const result: string[] = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmedLine = line.trim();

    // 检查是否是代码块标记行
    if (trimmedLine.startsWith('```')) {
      // 如果只是 ```（没有语言标识符），这是代码块的闭合标记，应该保留
      if (trimmedLine === '```') {
        result.push(line);
        continue;
      }

      // 提取语言标识符
      const language = trimmedLine.substring(3).trim().toLowerCase();

      // 检查语言标识符是否有效
      if (!language || !validLanguageSet.has(language)) {
        // 无效的语言标识符，移除整行
        continue;
      }
    }

    // 有效行或非代码块标记行，保留
    result.push(line);
  }

  return result.join('\n');
}
function removeInvalidCode1(markdownStr: string): string {
  // 将语言列表转换为Set，便于快速查找
  const validLanguageSet = new Set(validLanguages);

  // 按行分割字符串
  const lines = markdownStr.split('\n');
  const result: string[] = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmedLine = line.trim();

    // 检查是否是代码块标记行
    if (trimmedLine.startsWith('~~~')) {
      // 如果只是 ~~~（没有语言标识符）
      if (trimmedLine === '~~~') {
        // 这是无效的，跳过这行（不添加到结果中）
        continue;
      }

      // 提取语言标识符
      const language = trimmedLine.substring(3).trim().toLowerCase();

      // 检查语言标识符是否有效
      if (!language || !validLanguageSet.has(language)) {
        // 无效的语言标识符，移除整行
        continue;
      }
    }

    // 有效行或非代码块标记行，保留
    result.push(line);
  }

  return result.join('\n');
}

/** 过滤代码为空的情况 */
function filterEmptyCodeBlocks(markdownString: string) {
  // 创建语言集合用于快速查找
  const languageSet = new Set(validLanguages);

  // 处理不完整的代码块（只有开始标记，在字符串末尾）
  // 匹配 ```language 后面跟着换行符或空白字符直到字符串结束
  let result = markdownString.replace(/```([a-zA-Z0-9_+-]*)\s*$/g, (match, language) => {
    const langLower = language.toLowerCase();

    if (langLower === '' || languageSet.has(langLower)) {
      return '';
    }
    return match;
  });

  // 处理完整的代码块
  result = result.replace(/```([a-zA-Z0-9_+-]*)\n([\s\S]*?)\n```/g, (match, language, content) => {
    const langLower = language.toLowerCase();

    if (langLower === '' || languageSet.has(langLower)) {
      const trimmedContent = content.replace(/\s/g, '');
      if (trimmedContent === '') {
        return '';
      }
    }
    return match;
  });

  // 处理直接连接的情况: ```lang```
  result = result.replace(/```([a-zA-Z0-9_+-]*)```/g, (match, language) => {
    const langLower = language.toLowerCase();

    if (langLower === '' || languageSet.has(langLower)) {
      return '';
    }
    return match;
  });

  // 处理只有换行没有内容的情况: ```lang\n```（但不在末尾）
  result = result.replace(/```([a-zA-Z0-9_+-]*)\n```/g, (match, language) => {
    const langLower = language.toLowerCase();

    if (langLower === '' || languageSet.has(langLower)) {
      return '';
    }
    return match;
  });

  return result;
}

/**
 * 移除不符合要求的 Markdown 代码块语法
 * @param markdownStr Markdown 字符串
 * @returns 处理后的字符串
 */
export const removeInvalidCodeBlocks = (markdownStr: string, filterEmptyCode: boolean = false): string => {
  if (filterEmptyCode) {
    return filterEmptyCodeBlocks(removeInvalidCode(removeInvalidCode1(markdownStr)));
  }
  // 替换不符合要求的代码块开始标记
  return removeInvalidCode(removeInvalidCode1(markdownStr));
};

/**
 * 移除孤立的代码块结束标记
 * @param str 字符串
 * @returns 处理后的字符串
 */
export const removeOrphanedClosingBlocks = (str: string): string => {
  const lines = str.split('\n');
  const result: string[] = [];
  let inValidCodeBlock = false;

  for (const line of lines) {
    const trimmedLine = line.trim();

    // 检查是否是有效的代码块开始
    if (trimmedLine.startsWith('```') && trimmedLine.length > 3) {
      const language = trimmedLine.substring(3).trim();
      if (language) {
        inValidCodeBlock = true;
        result.push(line);
        continue;
      }
    }

    // 检查是否是代码块结束
    if (trimmedLine === '```') {
      if (inValidCodeBlock) {
        inValidCodeBlock = false;
        result.push(line);
      }
      // 如果不在有效代码块中，跳过这个结束标记
      continue;
    }

    result.push(line);
  }

  return result.join('\n');
};
