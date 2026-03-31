import { del, post, put, get } from '@/utils/http';
import type { AddFileToTempParams, CreateTempParams, CreateTempResponseSourceType, TempFileType } from './interface';
import { message } from 'antd';

const apiBaseUrl = '/api/agent-factory/v1';
const tempBaseUrl = `${apiBaseUrl}/temparea`;

// 获取Agent Api文档
export const getAPIDoc = ({
  app_key,
  agent_id,
  agent_version,
}: {
  // agent app key(通用agent传入 agent id，超级助手传入固定字符串:super_assistant)
  app_key: string;
  agent_id: string;
  agent_version: string;
}) =>
  post(`${apiBaseUrl}/app/${app_key}/api/doc`, {
    body: {
      agent_id,
      agent_version,
    },
  });

type FileType = {
  id: string;
  type: string;
};

type CheckFileStatusRes = {
  progress: number;
  process_info: Array<{ id: string; status: 'processing' | 'completed' | 'failed' }>;
};

/* 判断文件索引状态 */
export async function checkFileStatus(data: FileType[]): Promise<CheckFileStatusRes> {
  try {
    const res = await post(`${apiBaseUrl}/file/check`, { body: data });
    return res;
  } catch (error) {
    console.log(error);
    message.error('临时区文件索引失败');
    return { progress: 0, process_info: [] };
  }
}

/* 创建临时区 */
export async function createTemp(data: CreateTempParams) {
  try {
    const res = await post(tempBaseUrl, { body: data });
    return res;
  } catch (error) {
    console.log(error);
    message.error('临时区创建失败');
    return false;
  }
}

/* 临时区添加文件 */
export async function addFileToTemp({
  id,
  ...data
}: AddFileToTempParams): Promise<CreateTempResponseSourceType[] | false> {
  try {
    const res = await put(`${tempBaseUrl}/${id}`, { body: data });
    return res;
  } catch (error) {
    console.log(error);
    message.error('临时区文件添加失败');
    return false;
  }
}

/* 删除临时区文件 */
export async function removeTempFile({ tempId, sourceIds }: { tempId: string; sourceIds: string[] }) {
  try {
    const res = await del(`${tempBaseUrl}/${tempId}?${sourceIds.map(id => `source_id=${id}`).join('&')}`);
    return res || true;
  } catch (error) {
    console.log(error);
    message.error('临时区文件删除失败');
    return false;
  }
}

/* 获取临时区的文件 */
export async function getTempFiles(id: string): Promise<TempFileType[] | false> {
  try {
    const res = await get(`${tempBaseUrl}/${id}`);
    return res;
  } catch (error) {
    console.log(error);
    message.error('临时区文件获取失败');
    return false;
  }
}

// 获取Agent可观测指标数据
export const getAgentObservabilityMetrics = async (
  agentId: string,
  params: {
    agent_version: string;
    start_time: number;
    end_time: number;
  }
) => {
  try {
    const res = await post(`${apiBaseUrl}/observability/agent/${agentId}/detail`, { body: params });
    return res;
  } catch (error) {
    console.log(error);
    message.error('获取可观测数据失败');
    return false;
  }
};

// 获取Agent可观测页面的对话列表
export const getAgentObservabilityConversationList = async (
  agentKey: string,
  params: {
    agent_id: string;
    agent_version: string;
    start_time: number;
    end_time: number;
    title: string;
    page: number;
    size: number;
  }
) => {
  try {
    const res = await post(`${apiBaseUrl}/observability/agent/${agentKey}/conversation`, { body: params });
    return res;
  } catch (error) {
    console.log(error);
    message.error('获取可观测数据失败');
    return false;
  }
};

// 获取Agent可观测页面对话的session列表
export const getAgentObservabilitySessionList = async (
  agentId: string,
  conversationId: string,
  params: {
    agent_id: string;
    agent_version: string;
    start_time: number;
    end_time: number;
    page: number;
    size: number;
  }
) => {
  try {
    const res = await post(`${apiBaseUrl}/observability/agent/${agentId}/conversation/${conversationId}/session`, {
      body: params,
    });
    return res;
  } catch (error) {
    console.log(error);
    message.error('获取可观测数据失败');
    return false;
  }
};

// // 获取Agent可观测页面对话的session 指标
export const getAgentObservabilitySessionMetric = async (
  agentId: string,
  conversationId: string,
  sessionId: string,
  params: {
    agent_id: string;
    agent_version: string;
    start_time: number;
    end_time: number;
  }
) => {
  try {
    const res = await post(
      `${apiBaseUrl}/observability/agent/${agentId}/conversation/${conversationId}/session/${sessionId}/detail`,
      {
        body: params,
      }
    );
    return res;
  } catch (error) {
    console.log(error);
    message.error('获取可观测数据失败');
    return false;
  }
};

// 获取Agent可观测页面的对话的session的run列表
export const getAgentObservabilityRunList = async (
  agentId: string,
  conversationId: string,
  sessionId: string,
  params: {
    agent_id: string;
    agent_version: string;
    start_time: number;
    end_time: number;
    page: number;
    size: number;
  }
) => {
  try {
    const res = await post(
      `${apiBaseUrl}/observability/agent/${agentId}/conversation/${conversationId}/session/${sessionId}/run`,
      {
        body: params,
      }
    );
    return res;
  } catch (error) {
    console.log(error);
    message.error('获取可观测数据失败');
    return false;
  }
};

// 获取Agent可观测页面的对话的session的run详情
export const getAgentObservabilityRunDetail = async (
  agentId: string,
  conversationId: string,
  sessionId: string,
  runId: string,
  params: {
    agent_id: string;
    agent_version: string;
    start_time: number;
    end_time: number;
  }
) => {
  try {
    const res = await post(
      `${apiBaseUrl}/observability/agent/${agentId}/conversation/${conversationId}/session/${sessionId}/run/${runId}/detail`,
      {
        body: params,
      }
    );
    return res;
  } catch (error) {
    console.log(error);
    message.error('获取可观测数据失败');
    return false;
  }
};
