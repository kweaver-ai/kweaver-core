import React, { ReactNode, UIEventHandler } from 'react';
import type { TableProps } from 'antd';
import { TableLocale } from 'antd/lib/table/interface';

export type DisplayTypeProps = 'table' | 'card';

export type ButtonConfigProps = {
  key: string;
  label?: string;
  type?: 'add' | 'add-down' | 'delete' | 'filter' | 'order' | 'fresh' | 'switch' | 'upload' | string;
  position: 'left' | 'right';
  onHandle?: Function;
  // type=order 排序菜单相关配置
  orderMenu?: { id: string; intlText: string }[] | ReactNode; // 排序菜单列表
  orderField?: string; // 选中的排序字段
  order?: 'asc' | 'desc' | string; // 排序规则，asc 正序、desc 倒序
  onOrderMenuClick?: (params: any) => void;

  tip?: boolean;
  itemDom?: ReactNode;
  visible?: boolean;
  disabled?: boolean;
};

export type FilterToolsOptionProps = {
  id: string | number;
  label?: string;
  optionList?: { key: string | number; value: string; text: string }[];
  onHandle?: Function;
  value?: any; // 配置value 则代表受控
  itemDom?: ReactNode;
  showSearch?: boolean;
};

export type CardItemMenuItemType = {
  label: string | ReactNode | ((record: any) => ReactNode);
  disabled?: (record: any) => boolean;
  visible?: (record: any) => boolean;
  onClick: (record: any) => void;
};

export type CardItemConfigType = {
  field?: string;
  render?: (record: any) => ReactNode;
  type: 'text' | 'multiText' | 'icon' | 'status' | 'menu' | 'title' | 'time';
  menu?: CardItemMenuItemType[];
};

export type CardItemType = {
  header?: CardItemConfigType[];
  body?: CardItemConfigType[];
  footer?: CardItemConfigType[];
};

export type CardItemFuncType = (record: any, index: number) => ReactNode;

export type CardProps = {
  cardHeight?: number; // 卡片高度
  rowCardNum?: number; // 一行卡片的最大数量
  renderCardItem: CardItemType | CardItemFuncType; // 渲染卡片模式下的子元素
  gap?: number; // 卡片之间的间距
  onClick?: (record: any) => void;
};

export interface TitleProps {
  children?: ReactNode;
  visible?: boolean;
  title?: any;
  className?: string;
}

export interface HeaderProps {
  children?: ReactNode;
  title?: any;
  className?: string;
  visible?: boolean;
  showFilter?: boolean;
  onFilterClick?: Function;
  filterConfig?: any;
  renderButtonConfig?: ButtonConfigProps[];
  searchValue?: string;
  onSearchChange?: Function;
  searchPlaceholder?: string;
  filterToolsOptions?: FilterToolsOptionProps[];
  onDisplayChange?: (type: DisplayTypeProps) => void;
  displayType?: DisplayTypeProps;
  showSearch?: boolean;
}

export interface FilterOperationContainerProps {
  children?: ReactNode;
  visible?: boolean;
  showSearch?: boolean;
  className?: string;
  filterConfig?: any;
  filterToolsOptions?: FilterToolsOptionProps[];
  searchValue?: string;
  onSearchChange?: Function;
  searchPlaceholder?: string;
  onClose?: Function;
}

export interface ITableProps extends Omit<TableProps<any>, 'locale'> {
  width?: number | string;
  contextMenu?: {
    headerContextMenu: React.ReactElement<any>;
    bodyContextMenu: React.ReactElement<any>;
  };
  emptySvg?: any;
  emptyText?: string | ReactNode;
  lastColWidth?: number;
  persistenceID?: string | number;
  autoScrollY?: boolean; // 是否自动计算scrollY的值 默认false 以及自己配置scrollY的时候不会自动计算

  // 用于判断是否展示搜索/过滤不到数据情况下的缺醒图
  filterToolsOptions?: FilterToolsOptionProps[];
  searchValue?: string;
  locale?: TableLocale | null; // 传null 则代表不显示默认的缺省图
  onScrollCapture?: UIEventHandler<HTMLDivElement>;
  activeRowKey?: string; // 处于激活状态的行（列表页跳转详情，再跳回来，需要体现跳转的行）
}

export interface ADTableProps extends Omit<ITableProps, 'title'> {
  className?: string;
  title?: ReactNode;
  showFilter?: boolean;
  showSearch?: boolean; // 是否显示搜索，默认显示
  showTableHeader?: boolean; // 是否表格的表头（对应antd 原生Table的 showHeader属性）
  onFilterClick?: Function;
  renderButtonConfig?: ButtonConfigProps[];
  searchValue?: string;
  onSearchChange?: Function;
  searchPlaceholder?: string;
  filterToolsOptions?: FilterToolsOptionProps[];
  onFiltersToolsClose?: Function;
  card?: CardProps; // 卡片模式下的配置项
  defaultDisplayType?: DisplayTypeProps; // 默认显示类型 -  table
  extraContent?: ReactNode;
  customBody?: ReactNode;
  onDisplayTypeChange?: (type: 'card' | 'table') => void; // 列表展现形式的值变化事件,
  onScrollCapture?: UIEventHandler<HTMLDivElement>;
}
