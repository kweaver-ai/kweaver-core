export const getBlockIcon = (type: string) => {
  let icon = '';
  if (type === 'input') {
    icon = 'icon-agent-input';
  }
  if (type === 'skill') {
    icon = 'icon-agent-retrievers';
  }
  if (type === 'llm') {
    icon = 'icon-agent--LLM';
  }
  if (type === 'function_block') {
    icon = 'icon-hanshuguanli';
  }
  if (type === 'output') {
    icon = 'icon-agent-output';
  }
  return icon;
};
