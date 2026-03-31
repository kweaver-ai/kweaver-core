/* 临时区文件类型 */
export enum TempFileTypeEnum {
  Doc = 'doc',
}

export interface TempFileType {
  /* 文档类型 */
  type: TempFileTypeEnum;
  /* 文档id */
  id: string;
  details?: {
    id: string;
    name?: string;
    title?: string;
    suffix?: string;
    [key: string]: string | number;
  };
}

export interface CreateTempParams {
  /* 上传的文档信息 */
  source: TempFileType[];
  agent_id: string;
  agent_version?: string;
}

export interface CreateTempResponseSourceType {
  id: string;
  details: {
    status: 'success' | '';
    message: string;
  }[];
}

export interface CreateTempResponse {
  id: string;
  sources: CreateTempResponseSourceType[];
}

export interface AddFileToTempParams {
  id: string;
  /* 上传的文档信息 */
  source: TempFileType[];
  agent_id: string;
  agent_version?: string;
}
