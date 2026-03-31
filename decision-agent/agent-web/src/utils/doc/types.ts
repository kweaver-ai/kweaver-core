import React, { type ForwardRefExoticComponent, type RefAttributes } from 'react';
import { type IconComponentProps } from '@ant-design/icons/lib/components/Icon';

export interface IFileIcons {
  [key: string]:
    | ForwardRefExoticComponent<IconComponentProps & RefAttributes<HTMLSpanElement>>
    | React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
}

export enum DocLibType {
  UserDocLib = 'user_doc_lib',
  DepartmentDocLib = 'department_doc_lib',
  CustomDocLib = 'custom_doc_lib',
  KnowledgeDocLib = 'knowledge_doc_lib',
}

export enum DocType {
  DocLib = 'doc_lib',
}

export enum DocLibTypeEnum {
  UserDocLib = 'user_doc_lib',
  SharedUserDocLib = 'shared_user_doc_lib',
  DepartmentDocLib = 'department_doc_lib',
  CustomDocLib = 'custom_doc_lib',
  KnowledgeDocLib = 'knowledge_doc_lib',
}
