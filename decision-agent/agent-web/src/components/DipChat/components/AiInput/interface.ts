import type { SuggestionProps, SenderProps } from '@ant-design/x';
import type { AgentAppType, FileItem } from '@/components/DipChat/interface';

export type SuggestionItems = SuggestionProps['items'];

export type AiInputMode = 'normal';

export type AiInputValue = {
  mode: AiInputMode; // 模式
  inputValue: string;
  deepThink: boolean;
};

export type SuggestionItem = {
  items: SuggestionItems;
  triggerChar: string;
};

export interface AiInputProps extends Omit<SenderProps, 'onSubmit' | 'value' | 'onChange'> {
  suggestions?: SuggestionItem[];
  onSubmit?: (value: AiInputValue) => void;
  clearAfterSend?: boolean; // 发送触发后 清空输入框的内容
  value: AiInputValue;
  onChange?: (value: AiInputValue) => void;
  agentConfig: any; // Agent的配置信息
  agentAppType: AgentAppType; // Agent应用类型
  tempFileList?: FileItem[];
  onPreviewFile?: (file: FileItem) => void;
}

export type AiInputRef = {
  reset: () => void; // 清空输入的文本 & 上传的文件
};
