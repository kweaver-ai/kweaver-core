// 基于toolbox.yaml的新工具箱类型定义
export interface ToolBoxListParams {
  page?: number;
  page_size?: number;
  status?: 'unpublish' | 'published' | 'offline';
  user_id?: string;
  box_category?: string;
  name?: string;
  sort_by?: 'create_time' | 'updated_time' | 'name';
  sort_order?: 'asc' | 'desc';
  all?: boolean;
}

export interface ToolBoxInfoNew {
  box_id: string;
  box_name: string;
  box_desc: string;
  box_svc_url: string;
  status: 'unpublish' | 'published' | 'offline';
  category_type: string;
  category_name: string;
  is_internal: boolean;
  source: string;
  tools: string[];
  create_time: number;
  create_user: string;
  update_time: number;
  update_user: string;
}

export interface ToolBoxListResponse {
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
  data: ToolBoxInfoNew[];
}

export interface ToolListParams {
  page?: number;
  page_size?: number;
  sort_by?: 'create_time' | 'update_time' | 'name';
  sort_order?: 'asc' | 'desc';
  name?: string;
  status?: 'enabled' | 'disabled';
  user_id?: string;
  all?: boolean;
}

export interface QuotaControl {
  quota_type: 'token' | 'api_key' | 'ip' | 'user' | 'concurrent' | 'rate_limit' | 'none';
  quota_value?: number;
  time_window?: {
    value: number;
    unit: 'second' | 'minute' | 'hour' | 'day';
  };
  overage_policy?: 'reject' | 'queue' | 'log_only';
  burst_capacity?: number;
}

export interface ParameterSchema {
  type: 'string' | 'number' | 'integer' | 'boolean' | 'array';
  format?: 'int32' | 'int64' | 'float' | 'double' | 'byte';
  example?: string;
}

export interface OpenAPIStruct {
  summary: string;
  path: string;
  method: string;
  description: string;
  server_url: string;
  api_spec: {
    parameters?: Array<{
      name: string;
      in: 'path' | 'query' | 'header' | 'cookie';
      description: string;
      required: boolean;
      schema: ParameterSchema;
    }>;
    request_body?: {
      description: string;
      required: boolean;
      content: Record<
        string,
        {
          schema: any;
          example: any;
        }
      >;
    };
    responses?: Record<
      string,
      {
        description: string;
        content: Record<
          string,
          {
            schema: any;
            example: any;
          }
        >;
      }
    >;
    schemas?: Record<string, ParameterSchema>;
    security?: Array<{
      securityScheme: 'apiKey' | 'http' | 'oauth2';
    }>;
  };
}

export interface ToolInfoNew {
  tool_id: string;
  name: string;
  description: string;
  status: 'enabled' | 'disabled';
  metadata_type: 'openapi';
  metadata: OpenAPIStruct;
  quota_control?: QuotaControl;
  create_time: number;
  create_user: string;
  update_time: number;
  update_user: string;
  extend_info?: Record<string, any>;
}

export interface BoxToolListResponse {
  box_id: string;
  status: 'unpublish' | 'published' | 'offline';
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
  tools: ToolInfoNew[];
}

// 全局工具列表查询参数
export interface GlobalToolListParams {
  page?: number;
  page_size?: number;
  sort_by?: 'create_time' | 'update_time' | 'name';
  sort_order?: 'asc' | 'desc';
  name?: string;
  status?: 'enabled' | 'disabled';
  user_id?: string;
  all?: boolean;
}

// 全局工具列表响应
export interface GlobalToolListResponse {
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
  data: ToolInfoNew[];
}

// 获取单个工具的响应类型
export interface ToolResponse {
  tool_id: string;
  name: string;
  description: string;
  status: 'enabled' | 'disabled';
  metadata_type: 'openapi';
  metadata: OpenAPIStruct;
  quota_control?: QuotaControl;
  create_time: number;
  create_user: string;
  update_time: number;
  update_user: string;
  extend_info?: Record<string, any>;
}

// 工具箱市场列表查询参数
export interface ToolBoxMarketListParams {
  box_ids: string[]; // 工具箱ID数组
  fields: string; // 需要返回的字段，如 'box_name,box_desc'
}

// 工具箱市场列表响应 - 单个工具箱信息
export interface ToolBoxMarketInfo {
  box_id: string;
  box_name: string;
  box_desc: string;
}

// 工具箱市场列表响应
export type ToolBoxMarketListResponse = ToolBoxMarketInfo[];
