// 和后端约定提示词被$包裹
export const getPromptVarFromString = (str: string) => {
  const regex = /\$([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)*)/g;
  const targetArr = str.match(regex) || [];
  // 排除系统保留的变量
  return targetArr.filter(item => !item.includes('1_tool_use') && !item.includes('2_format_constraint'));
};

export const getVarTextFromPromptVar = (str: string) => {
  // 如果是$格式的变量，直接返回相应的部分
  if (str.startsWith('$')) {
    return ['$', str.slice(1), ''];
  }

  // 兼容原有的{{}}格式
  const regex = /(\{\{|\}\})/g; // 全局搜索，捕获 '{{' 或 '}}'
  const parts = str.split(regex, -1); // -1 作为第二个参数，保持捕获组
  if (parts.length >= 3) {
    parts.shift();
    parts.pop();
    return parts;
  }

  // 如果都不匹配，返回原字符串
  return ['', str, ''];
};
