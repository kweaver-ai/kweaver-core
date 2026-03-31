import { get, businessDomainHeaderKey } from '@/utils/http';

// MCP服务基本信息类型
export interface MCPServerReleaseInfo {
  mcp_id: string;
  name: string;
  description: string;
  mode: string;
  source: string;
  category: string;
  create_user: string;
  create_time: number;
}

// MCP工具类型
export interface MCPTool {
  name: string;
  description: string;
  inputSchema: {
    type: string;
    properties: Record<string, any>;
    required?: string[];
  };
  annotations?: {
    title?: string;
    readOnlyHint?: boolean;
    destructiveHint?: boolean;
    idempotentHint?: boolean;
    openWorldHint?: boolean;
  };
}

// 分页响应基础类型
interface PaginationResponse {
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  has_next: boolean;
  has_pre: boolean;
}

// MCP服务列表响应
export interface MCPServerListResponse extends PaginationResponse {
  data: MCPServerReleaseInfo[];
}

// MCP服务详情响应
export interface MCPServerDetailResponse {
  base_info: MCPServerReleaseInfo;
  tools: MCPTool[];
}

// MCP代理工具列表响应
export interface MCPProxyToolListResponse {
  tools: MCPTool[];
}

// 查询参数类型
export interface MCPServerListParams {
  page?: number;
  page_size?: number;
  status?: 'unpublish' | 'published' | 'offline' | 'editing';
  mode?: 'stdio' | 'sse' | 'stream';
  name?: string;
  category?: string;
}

/**
 * 获取已发布的MCP服务列表
 */
export const getMCPServerList = (params: MCPServerListParams = {}, businessDomainIds?: string[]) => {
  return get('/api/agent-operator-integration/v1/mcp/market/list', {
    params: {
      page: 1,
      page_size: 100,
      status: 'published', // 默认只获取已发布的
      ...params,
    },
    headers: {
      [businessDomainHeaderKey]: businessDomainIds?.join(',') || '',
    },
  });
};

/**
 * 获取MCP服务市场详情信息
 */
export const getMCPServerDetail = (mcpId: string, businessDomainIds?: string[]) => {
  return get(`/api/agent-operator-integration/v1/mcp/market/${mcpId}`, {
    headers: {
      [businessDomainHeaderKey]: businessDomainIds?.join(',') || '',
    },
  });
};

/**
 * 获取指定MCP服务下的工具列表
 */
export const getMCPServerTools = (mcpId: string, businessDomainIds?: string[]) => {
  return get(`/api/agent-operator-integration/v1/mcp/proxy/${mcpId}/tools`, {
    headers: {
      [businessDomainHeaderKey]: businessDomainIds?.join(',') || '',
    },
  });
};
