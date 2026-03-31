import React from 'react';
import type { ButtonProps, InputProps, InputNumberProps, SelectProps } from 'antd';

type TooltipPlacement =
  | 'top'
  | 'left'
  | 'right'
  | 'bottom'
  | 'topLeft'
  | 'topRight'
  | 'bottomLeft'
  | 'bottomRight'
  | 'leftTop'
  | 'leftBottom'
  | 'rightTop'
  | 'rightBottom';
export type TextInterface = {
  tip?: React.ReactNode;
  intl?: string;
  level?: number;
  style?: React.CSSProperties;
  align?: string;
  strong?: number;
  noHeight?: boolean;
  tipPosition?: TooltipPlacement;
  className?: string;
  children?: React.ReactNode;
  block?: boolean;
  ellipsis?: boolean;
  subText?: boolean;
  title?: string;
  onClick?: React.MouseEventHandler<HTMLDivElement>;
};

declare const ButtonTypes: [
  'default',
  'primary',
  'ghost',
  'dashed',
  'link',
  'text',
  'text-b',
  'icon',
  'u-text',
  'u-icon',
  'u-icon-primary',
  'icon-text',
  'icon-text-link'
];
export type BtnType = (typeof ButtonTypes)[number];
type SizeTypes = 'small' | 'middle' | 'large' | 'smallest' | undefined;
export type ButtonType = Omit<ButtonProps, 'size' | 'type'> & {
  size?: SizeTypes;
  className?: string;
  children?: string | object;
  type?: BtnType;
  tip?: string;
  tipPosition?: TooltipPlacement;
};

export type ContainerType = Omit<SelectProps, ''> & {
  span?: string | number;
  className?: string;
  children?: string | object;
};

export type InputType = Omit<InputProps, ''> & {
  min?: any;
  max?: any;
  size?: string;
  className?: string;
  children?: string | object;
};

export type InputNumberType = Omit<InputNumberProps, ''> & {
  size?: string;
  className?: string;
};

export type SelectType = Omit<SelectProps, ''> & {
  size?: string;
  className?: string;
  children?: string | object;
};
