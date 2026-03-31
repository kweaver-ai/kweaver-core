import { Button, message, Upload } from 'antd';
import DipIcon from '@/components/DipIcon';
import { forwardRef, type ReactNode, useImperativeHandle } from 'react';
import type { UploadProps } from 'antd';
import { useDipChatStore } from '@/components/DipChat/store.tsx';
import { createConversation } from '@/apis/super-assistant';
import { getFileListFromSandBox, uploadFileToSandBox } from '@/apis/sandbox';

export type FileUploadBtnProps = {
  disabled?: boolean;
  customBtn?: ReactNode;
  onSuccess?: () => void;
};

export type FileUploadBtnRef = {
  getFileList: () => void;
  clearFileList: () => void;
};

const FileUploadBtn = forwardRef<FileUploadBtnRef, FileUploadBtnProps>((props, ref) => {
  const {
    dipChatStore: { agentDetails, agentAppKey, debug },
    getDipChatStore,
    setDipChatStore,
    getConversationData,
  } = useDipChatStore();
  const { disabled = false, customBtn } = props;
  const [messageApi, contextHolder] = message.useMessage();
  const sessionId = 'sess-agent-default';

  useImperativeHandle(ref, () => ({
    getFileList,
    clearFileList,
  }));

  const clearFileList = () => {
    setDipChatStore({ tempFileList: [] });
  };

  const getFileList = async (debugFilePath?: string) => {
    const conversationId = getDipChatStore().activeConversationKey;
    const path = `conversation-${conversationId}/uploads/temparea`;
    const res: any = await getFileListFromSandBox({
      sessionId,
      path,
      limit: 1000,
    });
    if (res) {
      let files = res.files;
      if (debugFilePath) {
        files = files.filter((item: any) => item.container_path.includes(debugFilePath));
      }
      const list = files.map((item: any) => {
        const fileName = item.name.split('/temparea/').pop();
        return {
          ...item,
          checked: debug,
          status: 'completed',
          name: fileName,
        };
      });
      setDipChatStore({ tempFileList: list });
    }
  };

  // 处理对话创建后的 store 和 URL 更新
  const handleConversation = (conversation_id: string) => {
    const url = new URL(window.location.href);
    url.searchParams.set('conversation_id', conversation_id);
    // 使用 history API 更新 URL 而不刷新页面
    window.history.replaceState({}, '', url.toString());
    setDipChatStore({ activeConversationKey: conversation_id });
    getConversationData();
  };

  // 自定义上传逻辑
  const customRequest: UploadProps['customRequest'] = async options => {
    const { file, onSuccess, onError } = options;
    const uploadFile = file as File;

    try {
      let conversationId = getDipChatStore().activeConversationKey;
      // 1. 如果没有 activeConversationKey，先创建对话
      if (!conversationId) {
        const conversationRes = await createConversation(agentAppKey, {
          agent_id: agentDetails.id,
          agent_version: debug ? 'v0' : agentDetails.version,
          executor_version: 'v2',
        });

        if (!conversationRes) {
          throw new Error('创建对话失败');
        }

        conversationId = conversationRes.id;
        // 更新 URL 和 store
        if (!debug) {
          handleConversation(conversationId);
        } else {
          setDipChatStore({ activeConversationKey: conversationId });
        }
      }
      const filePath = `conversation-${conversationId}/uploads/temparea/${encodeURIComponent(uploadFile.name)}`;
      const res: any = await uploadFileToSandBox({
        file: uploadFile,
        sessionId,
        filePath,
      });
      if (res) {
        onSuccess?.(res);
        if (debug) {
          getFileList?.(res.file_path);
        } else {
          getFileList?.();
        }
        props.onSuccess?.();
      }
    } catch (error: any) {
      messageApi.error(error.message || '上传失败');
      onError?.(error);
    }
  };

  const uploadProps: UploadProps = {
    customRequest,
    showUploadList: false,
    maxCount: 1,
    disabled,
  };

  return (
    <>
      {contextHolder}
      <Upload {...uploadProps}>{customBtn || <Button icon={<DipIcon type="icon-dip-attachment" />} />}</Upload>
    </>
  );
});
FileUploadBtn.displayName = 'FileUploadBtn';
export default FileUploadBtn;
