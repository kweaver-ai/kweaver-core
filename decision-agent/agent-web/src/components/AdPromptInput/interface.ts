import React, { type CSSProperties } from 'react';
import type { SelectProps } from 'antd';

export interface VarOptions {
  label: any;
  value: string;
  type: string;
}

export interface Position {
  x: number;
  y: number;
}

export type AdPromptInputProps = {
  className?: string;
  style?: CSSProperties;
  value?: string; // 提供可直接作为antd 的 FormItem 控制的表单
  onChange?: (inputContent: string) => void; // 提供可直接作为antd 的 FormItem 控制的表单
  placeholder?: string;
  maxLength?: number;
  onFocus?: () => void;
  onBlur?: () => void;
  bordered?: boolean; // 是否有边框
  disabled?: boolean;
  getPopupContainer?: SelectProps['getPopupContainer']; // 下拉选择框的挂载节点
  onKeyDown?: (e: React.KeyboardEvent<HTMLDivElement>) => void;
  footer?: React.ReactNode;
  trigger: Array<{ character: string; options: VarOptions[] }>;
};

export type AdPromptInputRef = {
  setValue: (value: string) => void; // 设置值
};
