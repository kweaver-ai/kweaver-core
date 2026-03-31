export const extractCustomVarFromBeforeCursor = (value: string) => {
  if (value === '') {
    return [];
  }
  const customVar: any = [];

  const lines = value.split('\n').filter(line => !/^\s*#/.test(line)); // 如果某一行以#开头（#前面可以有多个空格），就算被注释了，此行忽略
  lines.forEach(line => {
    const lineContent = line;

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
  });
  return customVar;
};
