export enum LangType {
  zh = 'zh-cn',
  tw = 'zh-tw',
  us = 'en-us',
}

export interface ConfigType {
  protocol: string;
  host: string;
  port: number;
  lang: LangType;
  prefix: string;
  getToken: () => string;
  refreshToken?: () => Promise<{ access_token: string }>;
  onTokenExpired?: (code?: number) => void;
  toast?: {
    warning: (msg: string) => void;
  };
  businessDomainID: string;
}

export interface OptionsType {
  body?: any;
  headers?: any;
  timeout?: number;
  params?: Record<string, any>;
  resHeader?: boolean;
  returnFullResponse?: boolean;
}

export const config: ConfigType = {
  protocol: 'http',
  host: 'localhost',
  port: 80,
  lang: LangType.zh,
  prefix: '',
  getToken: () => '',
  refreshToken: undefined,
  onTokenExpired: undefined,
  toast: undefined,
  businessDomainID: '',
};

export enum IncrementalActionEnum {
  Upsert = 'upsert',
  Append = 'append',
  Remove = 'remove',
  End = 'end',
}

export const businessDomainHeaderKey = 'x-business-domain';
