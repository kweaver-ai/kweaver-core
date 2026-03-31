import intl from 'react-intl-universal';
import { dolphinKeywords } from '../static';
// import { extractToolBoxFromImports, extractToolsFromImports } from '@/components/AdDolphinEditor/utils';
// import { dolphinAutoClosingPairs } from '@/components/AdMonacoEditor/dolphin/registerDolphinLanguage';

export const validateDolphinLanguageError = (value: string, monaco: any) => {
  let errorMarkers: any = [];
  if (value) {
    // let toolNames: any = {}; // 储存所有的工具名
    // let toolBoxNames: any = {}; // 储存所有的工具箱名
    const lines = value.split('\n');
    lines.forEach((line, lineNumber) => {
      const variableMatches = line.match(/(?:->|>>)\s*[^\s]+/g);
      if (variableMatches) {
        variableMatches.forEach(variable => {
          const variableName = variable.replace(/^(?:->|>>)\s*/, '');
          const keywords = dolphinKeywords.map(keyword => keyword.label);
          const markerField = {
            severity: monaco.MarkerSeverity.Error,
            startLineNumber: lineNumber + 1,
            startColumn: line.indexOf(variable) + 1,
            endLineNumber: lineNumber + 1,
            endColumn: line.indexOf(variable) + variable.length + 1
          };

          // 检测是否为关键字
          if (keywords.includes(variableName)) {
            errorMarkers.push({
              ...markerField,
              message: `${variableName} ${intl.get('agentCommonConfig.llm.dolphinVarError')}`
            });
          }
          // 检测变量名是否合法（不能以数字开头，只能包含字母数字下划线）
          else if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(variableName)) {
            errorMarkers.push({
              ...markerField,
              message: `${intl.get('agentCommonConfig.llm.dolphinVarNameError')}`
            });
          }
        });
      }
      // // 收集所有行的tool信息
      // if (line.trim().startsWith('import')) {
      //   extractToolsFromImports(line).forEach(item => {
      //     const [toolBoxName, toolName] = item.split('.');
      //     if (toolNames[toolName]) {
      //       toolNames[toolName].push({
      //         startLineNumber: lineNumber + 1,
      //         startColumn: line.indexOf(item) + 1,
      //         endLineNumber: lineNumber + 1,
      //         endColumn: line.indexOf(item) + item.length + 1
      //       });
      //     } else {
      //       toolNames[toolName] = [
      //         {
      //           startLineNumber: lineNumber + 1,
      //           startColumn: line.indexOf(item) + 1,
      //           endLineNumber: lineNumber + 1,
      //           endColumn: line.indexOf(item) + item.length + 1
      //         }
      //       ];
      //     }
      //   });
      //   extractToolBoxFromImports(line).forEach(toolBoxName => {
      //     if (toolBoxNames[toolBoxName]) {
      //       toolBoxNames[toolBoxName].push({
      //         startLineNumber: lineNumber + 1,
      //         startColumn: line.indexOf(toolBoxName) + 1,
      //         endLineNumber: lineNumber + 1,
      //         endColumn: line.indexOf(toolBoxName) + toolBoxName.length + 1
      //       });
      //     } else {
      //       toolBoxNames[toolBoxName] = [
      //         {
      //           startLineNumber: lineNumber + 1,
      //           startColumn: line.indexOf(toolBoxName) + 1,
      //           endLineNumber: lineNumber + 1,
      //           endColumn: line.indexOf(toolBoxName) + toolBoxName.length + 1
      //         }
      //       ];
      //     }
      //   });
      // }
    });
    // // 校验import的工具会不会重复, 并收集所有import的工具
    // const allToolNames = Object.keys(toolNames);
    // allToolNames.forEach(toolName => {
    //   if (toolNames[toolName].length > 1) {
    //     toolNames[toolName].forEach((item: any) => {
    //       errorMarkers.push({
    //         ...item,
    //         severity: monaco.MarkerSeverity.Error,
    //         message: intl.get('global.repeat', { name: toolName })
    //       });
    //     });
    //   }
    // });
    // // 校验import的工具箱会不会重复, 并收集所有import的工具箱
    // const allToolBoxNames = Object.keys(toolBoxNames);
    // allToolBoxNames.forEach(toolBoxName => {
    //   if (toolBoxNames[toolBoxName].length > 1) {
    //     toolBoxNames[toolBoxName].forEach((item: any) => {
    //       errorMarkers.push({
    //         ...item,
    //         severity: monaco.MarkerSeverity.Error,
    //         message: intl.get('global.repeat', { name: toolBoxName })
    //       });
    //     });
    //   }
    // });
    // // 遍历lines，取出每一行@方式和tools=[]调用的工具名称，判断调用的工具在allToolNames中是否存在
    // lines.forEach((line, lineNumber) => {
    //   // 检查@方式工具名称
    //   const funcMatches = line.match(/@[^\s(]+\([^)]*\)/g);
    //   if (funcMatches) {
    //     funcMatches.forEach(funcCall => {
    //       // 提取函数名（去掉@和括号部分）
    //       const funcName = funcCall.match(/@([^\s(]+)\(/)?.[1];
    //       if (funcName && !allToolNames.includes(funcName)) {
    //         errorMarkers.push({
    //           severity: monaco.MarkerSeverity.Error,
    //           message: intl.get('agentCommonConfig.llm.undefinedToolError', { name: funcName }),
    //           startLineNumber: lineNumber + 1,
    //           startColumn: line.indexOf(funcCall) + 1,
    //           endLineNumber: lineNumber + 1,
    //           endColumn: line.indexOf(funcCall) + funcCall.length + 1
    //         });
    //       }
    //     });
    //   }
    //   // 检查tools=[]中的工具名称
    //   const toolsMatch = line.match(/tools=\[(.*?)\]/);
    //   if (toolsMatch) {
    //     const toolsContent = toolsMatch[1].trim();
    //     if (toolsContent) {
    //       const toolsList = toolsContent.split(/\s*,\s*/); // 分割并去除空白字符
    //       toolsList.forEach(toolName => {
    //         if (toolName && !allToolNames.includes(toolName)) {
    //           const toolStart = line.indexOf(toolName, line.indexOf('tools=['));
    //           errorMarkers.push({
    //             severity: monaco.MarkerSeverity.Error,
    //             message: intl.get('agentCommonConfig.llm.undefinedToolError', { name: toolName }),
    //             startLineNumber: lineNumber + 1,
    //             startColumn: toolStart + 1,
    //             endLineNumber: lineNumber + 1,
    //             endColumn: toolStart + toolName.length + 1
    //           });
    //         }
    //       });
    //     }
    //   }
    // });

    // // 校验闭合标签
    // lines.forEach((line, lineNumber) => {
    //   const stack: { char: string; column: number }[] = [];
    //   let inString = false; // 是否在字符串内
    //   let stringChar = ''; // 当前字符串的引号类型
    //
    //   for (let i = 0; i < line.length; i++) {
    //     const currentChar = line[i];
    //
    //     // 处理转义字符
    //     if (currentChar === '\\') {
    //       i++; // 跳过下一个字符
    //       continue;
    //     }
    //
    //     // 处理字符串
    //     if (currentChar === '"' || currentChar === "'") {
    //       if (!inString) {
    //         // 不在字符串内，开始一个新的字符串
    //         inString = true;
    //         stringChar = currentChar;
    //         continue;
    //       } else if (currentChar === stringChar) {
    //         // 在字符串内，遇到相同的引号，结束字符串
    //         inString = false;
    //         stringChar = '';
    //         continue;
    //       }
    //       // 在字符串内，遇到不同的引号，作为普通字符处理
    //       continue;
    //     }
    //
    //     // 如果在字符串内，跳过其他字符的检查
    //     if (inString) {
    //       continue;
    //     }
    //
    //     // 检查是否是开始字符
    //     const pair = dolphinAutoClosingPairs.find(p => p.open === currentChar);
    //     if (pair) {
    //       stack.push({ char: pair.close, column: i });
    //       continue;
    //     }
    //
    //     // 检查是否是结束字符
    //     if (stack.length > 0 && currentChar === stack[stack.length - 1].char) {
    //       stack.pop();
    //     } else if (dolphinAutoClosingPairs.some(p => p.close === currentChar)) {
    //       // 发现结束字符但没有对应的开始字符
    //       errorMarkers.push({
    //         severity: monaco.MarkerSeverity.Error,
    //         message: intl.get('agentCommonConfig.llm.unmatchedClosingError'),
    //         startLineNumber: lineNumber + 1,
    //         startColumn: i + 1,
    //         endLineNumber: lineNumber + 1,
    //         endColumn: i + 2
    //       });
    //     }
    //   }
    //
    //   // 如果字符串没有正确闭合，添加错误标记
    //   if (inString) {
    //     const startColumn = line.lastIndexOf(stringChar) + 1;
    //     errorMarkers.push({
    //       severity: monaco.MarkerSeverity.Error,
    //       message: intl.get('agentCommonConfig.llm.unmatchedClosingError'),
    //       startLineNumber: lineNumber + 1,
    //       startColumn: startColumn,
    //       endLineNumber: lineNumber + 1,
    //       endColumn: startColumn + 1
    //     });
    //   }
    //
    //   // 检查未闭合的开始字符
    //   stack.forEach(item => {
    //     errorMarkers.push({
    //       severity: monaco.MarkerSeverity.Error,
    //       message: intl.get('agentCommonConfig.llm.unmatchedClosingError'),
    //       startLineNumber: lineNumber + 1,
    //       startColumn: item.column + 1,
    //       endLineNumber: lineNumber + 1,
    //       endColumn: item.column + 2
    //     });
    //   });
    // });
  }
  return errorMarkers;
};
