// 逻辑处理

import intl from 'react-intl-universal';
import type { DipChatState, DipChatItem } from './interface';
import { getAgentDetailInUsagePage } from '@/apis/agent-factory';
import { getChatItemContent } from './utils';
import _ from 'lodash';
import type { UseTypeOutResponse } from '@/hooks/useStreamingOut';
import { isJSONString } from '@/utils/handle-function';
import dayjs from 'dayjs';

/** 获取普通单一Agent配置基础信息 */
export const getCommentAgentConfig = ({
  agentId,
  agentVersion,
  customSpaceId,
}: any): Promise<Partial<DipChatState> | false> =>
  // eslint-disable-next-line no-async-promise-executor
  new Promise(async resolve => {
    const res: any = await getAgentDetailInUsagePage({
      id: agentId,
      version: agentVersion,
      is_visit: true,
      customSpaceId,
    });
    if (res) {
      resolve({
        agentDetails: res,
        agentAppKey: res.key,
      });
    } else {
      resolve(false);
    }
  });

/** 后端数据转化为前端渲染结构逻辑处理 */
export const handleChatItemContent = (
  newChatList: DipChatItem[],
  response: UseTypeOutResponse,
  debug: boolean = false
) => {
  const lastIndex = newChatList.length - 1;
  const contentObj = JSON.parse(response.content);
  console.log(contentObj, '通用场景处理');
  // 删除block_answer
  _.forEach(contentObj?.message?.content?.middle_answer?.progress, item => {
    if (item && 'block_answer' in item) {
      delete item.block_answer;
    }
  });
  const { user_message_id, assistant_message_id, message } = contentObj;
  const progress = _.get(message, 'content.middle_answer.progress') ?? [];
  if (progress.length === 0) {
    return;
  }
  console.log(progress, '++通用场景处理Progress++');
  newChatList[lastIndex].loading = response.pending;
  newChatList[lastIndex].generating = response.generating;
  if (assistant_message_id) {
    newChatList[lastIndex].key = assistant_message_id;
  }
  if (user_message_id) {
    newChatList[lastIndex - 1].key = user_message_id;
  }
  newChatList[lastIndex].content = _.get(message, 'content') ? getChatItemContent(message) : {};
  newChatList[lastIndex].interrupt = _.get(message, 'ext.interrupt_info') || {};
  newChatList[lastIndex].agentRunId = _.get(message, 'ext.agent_run_id');
  // debug模式下记录原始数据，方便调试区展示原始输出结果
  if (debug) {
    // 把之前的sourceData移除，只保留最后一次的sourceData
    for (let i = 0; i < newChatList.length; i++) {
      newChatList[i].sourceData = undefined;
    }
    newChatList[lastIndex].sourceData = contentObj;
  }
};

/** 处理流式输出过程中报的错误 */
export const handleStreamingError = (newChatList: DipChatItem[], response: UseTypeOutResponse, error: any) => {
  const contentObj = JSON.parse(response.content);
  const { user_message_id, assistant_message_id } = contentObj;
  const lastIndex = newChatList.length - 1;
  newChatList[lastIndex].loading = false;
  newChatList[lastIndex].generating = false;
  if (assistant_message_id) {
    newChatList[lastIndex].key = assistant_message_id;
  }
  if (user_message_id) {
    newChatList[lastIndex - 1].key = user_message_id;
  }
  console.log(error, '流式过程中报错了');
  // newChatList[lastIndex].error = _.get(error, ['BaseError', 'error_details']) || intl.get('dipChat.streamingError');
  newChatList[lastIndex].error = error;
};

/** 处理接口响应的过程中报的错误 */
export const handleResponseError = (newChatList: DipChatItem[], response: UseTypeOutResponse) => {
  console.log(response, '流式接口本身报错');
  const contentObj = isJSONString(response.content) ? JSON.parse(response.content) : {};
  console.log(contentObj, '流式接口本身报错信息解析');
  const { user_message_id, assistant_message_id } = contentObj;
  const lastIndex = newChatList.length - 1;
  newChatList[lastIndex].loading = false;
  newChatList[lastIndex].generating = false;
  if (assistant_message_id) {
    newChatList[lastIndex].key = assistant_message_id;
  }
  if (user_message_id) {
    newChatList[lastIndex - 1].key = user_message_id;
  }
  newChatList[lastIndex].error = response.error;
};

/** 处理会话分组 */
export const handleConversationGroup = (entries: any[]) => {
  // 将列表项按照时间分组
  const groupedItems: Record<string, any[]> = {
    [intl.get('dipChat.today')]: [],
    [intl.get('dipChat.within7Days')]: [],
    [intl.get('dipChat.within30Days')]: [],
  };

  // 先将每个会话按时间分组
  entries.forEach((listItem: any) => {
    const time = listItem.update_time ?? listItem.create_time;
    // const timestamp = Math.floor(time / 1000000); // 将纳秒转换为毫秒
    const timestamp = Math.floor(time); // 将纳秒转换为毫秒
    const date = dayjs(timestamp);
    const now = dayjs();
    let group = intl.get('dipChat.within30Days');

    if (date.isSame(now, 'day')) {
      group = intl.get('dipChat.today');
    } else if (date.isAfter(now.subtract(7, 'day'))) {
      group = intl.get('dipChat.within7Days');
    }

    // 添加到对应分组
    groupedItems[group].push({
      label: listItem.title,
      key: listItem.id,
      timestamp: time,
      status: listItem.status,
      unRead: listItem.message_index > listItem.read_message_index,
    });
  });

  // 构造最终的数组结构
  const items = [
    {
      label: intl.get('dipChat.today'),
      key: 'today',
      children: groupedItems[intl.get('dipChat.today')],
    },
    {
      label: intl.get('dipChat.within7Days'),
      key: 'within7days',
      children: groupedItems[intl.get('dipChat.within7Days')],
    },
    {
      label: intl.get('dipChat.within30Days'),
      key: 'within30days',
      children: groupedItems[intl.get('dipChat.within30Days')],
    },
  ];
  return items;
};
