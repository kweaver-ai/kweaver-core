import { get, post, del, put } from '@/utils/http';
import qs from 'qs';
import { message } from 'antd';
const agentAppV1BaseUrl = '/api/agent-factory/v1';
const agentFactoryV3BaseUrl = '/api/agent-factory/v3';

export const getChatUrl = (
  app_key: string,
  recover: boolean = false,
  debug: boolean = false
  // customSpaceId: string = ''
) => {
  if (debug) {
    return `${agentAppV1BaseUrl}/app/${app_key}/debug/completion`;
  }
  if (recover) {
    return `${agentAppV1BaseUrl}/app/${app_key}/chat/resume`;
  }
  // if (customSpaceId) {
  //   return `${agentAppV1BaseUrl}/app/${app_key}/chat/completion?custom_space_id=${customSpaceId}`;
  // }
  return `${agentAppV1BaseUrl}/app/${app_key}/chat/completion`;
};

export type GetConversationListOption = {
  page?: number;
  size?: number;
};

export const getConversationList = async (app_key: string, options: GetConversationListOption = {}) => {
  const newOptions = {
    page: 1,
    size: 1000,
    ...options,
  };
  try {
    const res = await get(`${agentAppV1BaseUrl}/app/${app_key}/conversation?${qs.stringify(newOptions)}`);
    return res;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

export const getConversationDetailsById = async (app_key: string, conversationId: string) => {
  try {
    const res = await get(`${agentAppV1BaseUrl}/app/${app_key}/conversation/${conversationId}`);
    return res;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

export const deleteConversationById = async (app_key: string, conversationId: string) => {
  try {
    const res = await del(`${agentAppV1BaseUrl}/app/${app_key}/conversation/${conversationId}`);
    return res || true;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

export const getBotIdByAgentKey = async (agentKey: string) =>
  new Promise<any>(resolve => {
    get(`${agentFactoryV3BaseUrl}/agent-market/agent/${agentKey}/version/latest`)
      .then((res: any) => {
        resolve(res);
      })
      .catch((error: any) => {
        console.error(error, '报错了');
        resolve(false);
      });
  });

type ConversationType = {
  title?: string;
  agent_id?: string;
  agent_version?: string;
  executor_version?: 'v1' | 'v2';
};

/**
 * 创建会话
 * 1。 chat接口请求之前调用
 * 2. 临时区文件上传之后需要创建一次
 * */
export const createConversation = async (app_key: string, body: ConversationType) => {
  try {
    const res = await post(`${agentAppV1BaseUrl}/app/${app_key}/conversation`, {
      body,
    });
    return res;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

/**
 * 编辑会话
 * 1。 更新标题
 * 2. 更新临时区id
 * */
export const updateConversation = async (app_key: string, conversation_id: string, body: ConversationType) => {
  try {
    const res = await put(`${agentAppV1BaseUrl}/app/${app_key}/conversation/${conversation_id}`, {
      body,
    });
    return res || true;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

export const stopConversation = async (
  app_key: string,
  conversationId: string,
  agent_run_id: string,
  interrupted_assistant_message_id: string
) => {
  try {
    const res = await post(`${agentAppV1BaseUrl}/app/${app_key}/chat/termination`, {
      body: {
        conversation_id: conversationId,
        agent_run_id,
        interrupted_assistant_message_id,
      },
    });
    return res || true;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

/**
 * 标记会话已读
 * */
export const markReadConversation = async (app_key: string, conversation_id: string, latest_read_index: number) => {
  try {
    const res = await put(`${agentAppV1BaseUrl}/app/${app_key}/conversation/${conversation_id}/mark_read`, {
      body: {
        latest_read_index,
      },
    });
    return res || true;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

/** 获取会话的session状态 */
export const getConversationSessionStatus = async (data: {
  agent_id: string;
  agent_version: string;
  conversation_id: string;
}) => {
  try {
    const res = await put(`${agentAppV1BaseUrl}/conversation/session/${data.conversation_id}`, {
      body: {
        agent_id: data.agent_id,
        agent_version: data.agent_version,
        action: 'get_info_or_create',
      },
    });
    return res;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

/** 恢复会话session */
export const recoverConversationSession = async (data: {
  agent_id: string;
  agent_version: string;
  conversation_id: string;
}) => {
  try {
    const res = await put(`${agentAppV1BaseUrl}/conversation/session/${data.conversation_id}`, {
      body: {
        agent_id: data.agent_id,
        agent_version: data.agent_version,
        action: 'recover_lifetime_or_create',
      },
    });
    return res;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};

/** 初始化已有会话的Session */
export const initConversationSession = async (data: {
  agent_id: string;
  agent_version: string;
  conversation_id: string;
}) => {
  try {
    const res = await post(`${agentAppV1BaseUrl}/conversation/session/init`, {
      body: data,
    });
    return res;
  } catch (error: any) {
    message.error(error.description);
    return false;
  }
};
