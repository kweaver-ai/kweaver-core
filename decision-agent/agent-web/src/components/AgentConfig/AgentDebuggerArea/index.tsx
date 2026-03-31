import DipChat from '@/components/DipChat';
import { useAgentConfig } from '../AgentConfigContext';

const AgentDebuggerArea = () => {
  const { state, actions } = useAgentConfig();
  // console.log(state, '最新的Agent配置');
  const saveAgent = (): Promise<boolean | string> =>
    new Promise(async resolve => {
      try {
        const result = await actions.saveAgent({ showSuccess: false });
        if (!result) {
          // 保存失败，由saveAgent内部处理错误消息
          console.log('保存失败');
          resolve(false);
        } else {
          resolve(result);
        }
      } catch (error) {
        console.error('保存Agent出错:', error);
        resolve(false);
      }
    });
  return <DipChat debug agentAppType="common" agentDetails={state} onSaveAgent={saveAgent} />;
};
export default AgentDebuggerArea;
