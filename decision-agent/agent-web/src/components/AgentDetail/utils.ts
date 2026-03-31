// 获取开场白
export function getRemark(agentInfo?: any): string {
  return (
    agentInfo?.config?.opening_remark_config?.fixed_opening_remark ||
    agentInfo?.config?.opening_remark_config?.dynamic_opening_remark_prompt
  );
}

// 获取预设问题
export function getPresetQuestions(agentInfo?: any): { question: string }[] {
  return agentInfo?.config?.preset_questions || [];
}
