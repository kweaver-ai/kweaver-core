// 获取光标之前的字符串
import intl from 'react-intl-universal';
import { dolphinKeywords } from '../static';

/** 判断，是不是在tools=[]内部 */
const isLastCommaInToolsBrackets = (str: string) => {
  // 找到最后一个逗号的位置
  const lastCommaIndex = str.lastIndexOf(',');
  if (lastCommaIndex === -1) return false;

  // 找到最后一个 tools=[ 的位置
  const toolsStart = str.lastIndexOf('tools=[');
  if (toolsStart === -1) return false;

  // tools=[ 后的第一个字符位置
  const bracketStart = toolsStart + 'tools=['.length;

  // 如果最后一个逗号在 tools=[ 之前，直接返回 false
  if (lastCommaIndex < bracketStart) return false;

  // 从 tools=[ 开始找到对应的右方括号
  let bracketLevel = 1;
  let matchedBracketEnd = -1;

  for (let i = bracketStart; i < str.length; i++) {
    const char = str[i];
    // 跳过字符串内的括号
    if (char === '"' || char === "'") {
      const quote = char;
      i++;
      while (i < str.length && str[i] !== quote) {
        if (str[i] === '\\') i++; // 跳过转义字符
        i++;
      }
      continue;
    }

    if (char === '[') {
      bracketLevel++;
    } else if (char === ']') {
      bracketLevel--;
      if (bracketLevel === 0) {
        matchedBracketEnd = i;
        break;
      }
    }
  }

  // 如果没找到匹配的右括号，返回 false
  if (matchedBracketEnd === -1) return false;

  // 判断最后一个逗号是否在 tools=[] 的方括号内
  return lastCommaIndex > bracketStart && lastCommaIndex < matchedBracketEnd;
};

// 获取光标之前的字符串
export const getCharBeforeCursor = (editor: any) => {
  const model = editor.getModel();
  const position = editor.getPosition();
  const lineContent = model.getLineContent(position.lineNumber);
  const beforeCursor = lineContent.substring(0, position.column - 1);
  return beforeCursor;
};

// 获取光标所在行的完整内容
export const getLineContentByCursor = (editor: any) => {
  const model = editor.getModel();
  const position = editor.getPosition();
  return model.getLineContent(position.lineNumber);
};

// 是否要调用之前块的变量
export const isTriggerVar = (value: string) => {
  return value.endsWith('$');
};

// 是否是调用工具
export const isToolCall = (value: string, lineContent: string = '') => {
  if (value.endsWith(', ')) {
    const lastCommaIndex = value.lastIndexOf(',');
    if (lastCommaIndex !== -1) {
      if (isLastCommaInToolsBrackets(lineContent)) {
        return true;
      }
    }
  }
  return value.endsWith('@') || value.endsWith('tools=');
};

// 是否是导入工具
export const isToolImport = (value: string) => {
  return value.endsWith('import ') || (value.startsWith('import ') && value.endsWith(', '));
};

// 根据指定字符，触发补全事件的执行 (所有注册的provideCompletionItems补全函数都会被执行)
export const triggerSuggestByChar = (editor: any) => {
  const beforeCursor = getCharBeforeCursor(editor);
  const lineContentCursor = getLineContentByCursor(editor);
  if (isToolCall(beforeCursor, lineContentCursor) || isTriggerVar(beforeCursor)) {
    editor.trigger('keyboard', 'editor.action.triggerSuggest', {});
  }
};

// // 根据指定字符，触发代码补全提示 (所有注册的provideCompletionItems补全函数都会被执行)
// export const triggerSuggestByChar = (char: string | string[], editor: any) => {
//   const chars = Array.isArray(char) ? char : [char];
//   const beforeCursor = getCharBeforeCursor(editor);
//   console.log(beforeCursor, '光标之前的字符串');
//   let trigger = false;
//   chars.forEach(char => {
//     if (beforeCursor.endsWith(char)) {
//       trigger = true;
//     }
//   });
//   if (trigger) {
//     editor.trigger('keyboard', 'editor.action.triggerSuggest', {});
//   }
// };

// 注册块变量补全监听事件
export const registerBlockVarCompletion = (monaco: any, options: any) => {
  return monaco.languages.registerCompletionItemProvider('dolphin', {
    provideCompletionItems: async (model: any, position: any) => {
      // 光标之前的字符串
      const charBeforeCursor = model.getValueInRange({
        startLineNumber: position.lineNumber,
        startColumn: 0,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      });
      let suggestions: any = [];
      if (options.length > 0 && isTriggerVar(charBeforeCursor)) {
        const word = model.getWordUntilPosition(position);
        const range = new monaco.Range(
          position.lineNumber,
          word?.startColumn,
          position.lineNumber,
          position.column // insert into current position
        );
        const suggestionsOptions = options.map((item: any, index: number) => {
          let detail = '';
          if (item.label.endsWith('.data')) {
            detail = intl.get('agentCommonConfig.llm.promptVarDataDesc');
          }
          if (item.label.endsWith('.text')) {
            detail = intl.get('agentCommonConfig.llm.promptVarTextDesc');
          }
          return {
            label: item.label,
            insertText: item.label,
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            detail: detail,
            range,
            sortText: String.fromCharCode(97 + index),
          };
        });
        suggestions = [...suggestions, ...suggestionsOptions];
      }

      return { suggestions };
    },
  });
};

// 注册import工具补全监听事件
export const registerImportToolCompletion = (monaco: any) => {
  return monaco.languages.registerCompletionItemProvider('dolphin', {
    provideCompletionItems: async (model: any, position: any) => {
      // 光标之前的字符串
      const charBeforeCursor = model.getValueInRange({
        startLineNumber: position.lineNumber,
        startColumn: 0,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      });
      let suggestions: any = [];
      if (isToolImport(charBeforeCursor)) {
        const word = model.getWordUntilPosition(position);
        const range = new monaco.Range(
          position.lineNumber,
          word?.startColumn,
          position.lineNumber,
          position.column // insert into current position
        );
        // const toolData: any = await getAllTools();
        // const toolBoxRes: any = await getAllToolBox();
        const toolData: any = [];
        const toolBoxRes: any = [];
        if (toolBoxRes && toolBoxRes.data.length > 0) {
          const toolBoxData = [...toolBoxRes.data, { box_name: 'Agent', box_id: 'built-in-agent' }];
          const options = toolBoxData.map((item: any, index: number) => {
            return {
              label: {
                label: item.box_name,
                description: item.box_id === 'built-in-agent' ? '' : intl.get('agentTool.toolBox'),
              },
              insertText: item.box_name,
              kind: monaco.languages.CompletionItemKind.Folder,
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              range,
            };
          });
          suggestions = [...suggestions, ...options];
        }
        if (toolData && toolData.length > 0) {
          const options = toolData.map((item: any, index: number) => {
            const insertText = `${item.tool_box_name}.${item.tool_name}`;
            return {
              // label: `${item.tool_name}(${item.tool_box_name})`,
              label: {
                label: `${item.tool_name}(${item.tool_box_name})`,
                description: item.tool_box_id === 'built-in-agent' ? 'Agent' : intl.get('agentTool.tool'),
              },
              // detail: item.tool_box_id === 'built-in-agent' ? 'Agent' : '工具',
              insertText,
              kind: monaco.languages.CompletionItemKind.Function,
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              range,
            };
          });
          suggestions = [...suggestions, ...options];
        }
      }

      return {
        suggestions: suggestions.map((item: any, index: number) => ({
          ...item,
          sortText: String.fromCharCode(97 + index),
        })),
      };
    },
  });
};

// 注册关键词补全监听事件
export const registerKeywordsCompletion = (monaco: any) => {
  // 添加静态关键字的智能提示
  return monaco.languages.registerCompletionItemProvider('dolphin', {
    provideCompletionItems: async (model: any, position: any) => {
      // 光标之前的字符串
      const charBeforeCursor = model.getValueInRange({
        startLineNumber: position.lineNumber,
        startColumn: 0,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      });
      const lineContent = model.getLineContent(position.lineNumber);
      let suggestions: any = [];
      if (!isTriggerVar(charBeforeCursor) && !isToolCall(charBeforeCursor, lineContent)) {
        const word = model.getWordUntilPosition(position);
        const range = new monaco.Range(
          position.lineNumber,
          word.word ? word?.startColumn : word?.startColumn - 1,
          position.lineNumber,
          position.column
        );
        suggestions = dolphinKeywords.map(keywordItem => ({
          ...keywordItem,
          kind: monaco.languages.CompletionItemKind.Keyword,
          insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          detail: '关键字',
          range,
        }));
      }
      return { suggestions };
    },
  });
};

/** 从dolphin语言的import语句中提取工具和工具箱*/
export const extractToolAndToolBoxFromImports = (str: string) => {
  if (str === '') {
    return [];
  }
  // 先按行分割
  const lines = str.split('\n');
  // 过滤出import语句并去掉import关键字
  const importLines = lines
    .filter(line => line.trim().startsWith('import'))
    .map(line => line.replace(/^import\s*/i, ''));

  // 从import语句中提取工具名称和工具箱名称
  const result: string[] = [];
  const processedToolBoxes = new Set<string>(); // 用于记录已处理过的工具箱

  importLines.forEach(line => {
    const lineResult: string[] = [];
    // 修改正则表达式以支持名称中的空格，并捕获分隔符
    const regex = /([\u4e00-\u9fa5a-zA-Z0-9_\s-]+?)(?:\.([\u4e00-\u9fa5a-zA-Z0-9_\s-]+?))?(?:,|$|\s|;)/g;
    let match;
    let lastIndex = 0;

    while ((match = regex.exec(line)) !== null) {
      const [fullMatch, toolBox, toolName] = match;
      const trimmedToolBox = toolBox.trim();

      if (toolName) {
        // 是工具
        const trimmedToolName = toolName.trim();
        lineResult.push(`${trimmedToolBox}.${trimmedToolName}`);
        processedToolBoxes.add(trimmedToolBox);
      } else if (!processedToolBoxes.has(trimmedToolBox)) {
        // 是工具箱，且未被处理过
        lineResult.push(trimmedToolBox);
        processedToolBoxes.add(trimmedToolBox);
      }
      lastIndex = match.index + fullMatch.length;
    }

    result.push(...lineResult);
  });

  return result;
};

/** 从dolphin语言的import语句中提取工具 */
export const extractToolsFromImports = (str: string) => {
  if (str === '') {
    return [];
  }
  // 先按行分割
  const lines = str.split('\n');
  // 过滤出import语句并去掉import关键字
  const importLines = lines
    .filter(line => line.trim().startsWith('import'))
    .map(line => line.replace(/^import\s*/i, ''));

  // 从import语句中提取工具名称
  const tools: string[] = [];
  // 修改正则表达式以支持名称中的空格
  const regex = /([\u4e00-\u9fa5a-zA-Z0-9_\s-]+?)\.([\u4e00-\u9fa5a-zA-Z0-9_\s-]+?)(?:,|$|\s|;)/g;

  importLines.forEach(line => {
    let match;
    while ((match = regex.exec(line)) !== null) {
      // 去除工具箱名称和工具名称首尾的空格
      const toolBox = match[1].trim();
      const toolName = match[2].trim();
      tools.push(`${toolBox}.${toolName}`);
    }
  });

  return tools;
};

/** 从dolphin语言的import语句中提取工具箱 */
export const extractToolBoxFromImports = (str: string) => {
  if (str === '') {
    return [];
  }
  // 先按行分割
  const lines = str.split('\n');
  // 过滤出import语句并去掉import关键字
  const importLines = lines
    .filter(line => line.trim().startsWith('import'))
    .map(line => line.replace(/^import\s*/i, ''));

  // 从import语句中提取工具名称
  const tools: string[] = [];
  // 修改正则表达式以支持名称中的空格
  const regex = /([\u4e00-\u9fa5a-zA-Z0-9_\s-]+?)(?:,|$|\s|;)/g;

  importLines.forEach(line => {
    let match;
    while ((match = regex.exec(line)) !== null) {
      // 去除工具箱名称和工具名称首尾的空格
      const toolBox = match[1].trim();
      tools.push(toolBox);
    }
  });

  return tools;
};

// 注册调用import进来的工具补全监听事件
export const registerCallToolCompletion = (monaco: any, tools: string[]) => {
  return monaco.languages.registerCompletionItemProvider('dolphin', {
    provideCompletionItems: async (model: any, position: any) => {
      // 光标之前的字符串
      const charBeforeCursor = model.getValueInRange({
        startLineNumber: position.lineNumber,
        startColumn: 0,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      });
      const lineContent = model.getLineContent(position.lineNumber);
      let suggestions: any = [];
      if (isToolCall(charBeforeCursor, lineContent)) {
        const word = model.getWordUntilPosition(position);
        const range = new monaco.Range(
          position.lineNumber,
          word?.startColumn,
          position.lineNumber,
          position.column // insert into current position
        );
        if (tools.length > 0) {
          const options = tools.map((item: any, index: number) => {
            const [toolBoxName, toolName] = item.split('.');
            let insertText = `${toolName}()`;
            if (charBeforeCursor.endsWith('tools=')) {
              // tools=["xxx"]，需要加引号
              insertText = `["${toolName}"]`;
            }
            if (charBeforeCursor.endsWith(', ')) {
              const lastCommaIndex = charBeforeCursor.lastIndexOf(',');
              if (lastCommaIndex !== -1) {
                if (isLastCommaInToolsBrackets(lineContent)) {
                  // tools=["...", "xxx"]，在方括号内新增一个，需要加引号
                  insertText = `"${toolName}"`;
                }
              }
            }
            return {
              label: `${toolName}(${toolBoxName})`,
              insertText,
              kind: monaco.languages.CompletionItemKind.Keyword,
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              range,
              sortText: String.fromCharCode(97 + index), // 保证补全项的顺序是接口返回的顺序
            };
          });
          suggestions = [...suggestions, ...options];
        }
      }

      return { suggestions };
    },
  });
};

// 从光标之前的代码中提取自定义的变量
export const extractCustomVarFromBeforeCursor = (value: string, editor: any) => {
  if (value === '') {
    return [];
  }
  const customVar: any = [];
  // const model = editor.getModel();
  const position = editor.getPosition();
  // const lineContent = model.getLineContent(position.lineNumber);
  // const beforeCursor = lineContent.substring(0, position.column - 1);
  // 先按行分割
  const lines = value.split('\n').filter(line => !/^\s*#/.test(line)); // 如果某一行以#开头（#前面可以有多个空格），就算被注释了，此行忽略
  lines.forEach((line, index: number) => {
    let lineContent = line;
    if (position.lineNumber >= index + 1) {
      if (position.lineNumber === index + 1) {
        lineContent = lineContent.substring(0, position.column - 1);
      }
      const variableMatches = lineContent.match(/(?:->|>>)\s*[^\s]+/g);
      if (variableMatches) {
        variableMatches.forEach(variable => {
          const variableName = variable.replace(/^(?:->|>>)\s*/, '');
          customVar.push({
            label: variableName,
            value: variableName,
          });
        });
      }
    }
  });
  return customVar;
};
