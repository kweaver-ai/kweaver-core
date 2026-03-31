import DipChat from '@/components/DipChat';
import { useMicroWidgetProps } from '@/hooks';
import { useEffect, useMemo } from 'react';
import { getParam } from '@/utils/handle-function';
import { useLocation } from 'react-router-dom';
import type { AiInputValue } from '@/components/DipChat/components/AiInput/interface';
import _ from 'lodash';
import type { DipChatItem } from '@/components/DipChat/interface';
import { nanoid } from 'nanoid';

const AgentUsage = () => {
  const location = useLocation();
  const microWidgetProps = useMicroWidgetProps();
  useEffect(() => {
    microWidgetProps?.toggleSideBarShow?.(false);
    return () => {
      microWidgetProps?.toggleSideBarShow?.(true);
    };
  }, []);
  const agentInfo = useMemo(() => {
    // 说明是其他微应用通过navigate 跳转进入Agent使用界面
    let aiInputValue = location.state as AiInputValue;
    // 如果 location.state 为空，尝试从 URL 参数中获取 state 数据
    if (_.isEmpty(aiInputValue)) {
      const stateParam = getParam('state');
      if (stateParam) {
        try {
          aiInputValue = JSON.parse(decodeURIComponent(stateParam));
          const url = new URL(window.location.href);
          url.searchParams.delete('state');
          window.history.pushState({}, '', url.toString());
        } catch {
          // 忽略解析错误，使用默认值
        }
      }
    }

    const objData: any = {
      agentId: getParam('id'),
      agentVersion: getParam('version'),
      agentAppType: getParam('agentAppType'),
    };
    if (!_.isEmpty(aiInputValue)) {
      const cloneChatList: DipChatItem[] = [];
      cloneChatList.push({
        key: nanoid(),
        role: 'user',
        content: aiInputValue.inputValue,
        loading: false,
        fileList: aiInputValue.fileList,
      });
      cloneChatList.push({
        key: nanoid(),
        role: 'common',
        content: '',
        loading: true,
      });
      objData.defaultChatList = cloneChatList;
      objData.defaultAiInputValue = aiInputValue;
    }
    return objData;
  }, []);
  return <DipChat {...agentInfo} />;
};

export default AgentUsage;
