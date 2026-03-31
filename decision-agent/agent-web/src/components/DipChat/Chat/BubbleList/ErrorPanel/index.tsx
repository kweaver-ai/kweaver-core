import React, { useState } from 'react';
import { useDipChatStore } from '@/components/DipChat/store';
import { FileSearchOutlined } from '@ant-design/icons';
import { Button, Popover, Tooltip } from 'antd';
import DipModal from '@/components/DipModal';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';
import _ from 'lodash';
import DipIcon from '@/components/DipIcon';
import PanelFooter from '@/components/DipChat/Chat/BubbleList/PanelFooter';
import AgentIcon from '@/components/AgentIcon';
import DipButton from '@/components/DipButton';
import { useMicroWidgetProps } from '@/hooks';

const ErrorPanel = ({ chatItemIndex, readOnly }: any) => {
  const {
    dipChatStore: { chatList, logQueryAgentDetails, agentDetails, debug },
  } = useDipChatStore();
  const microWidgetProps = useMicroWidgetProps();
  const chatItem = chatList[chatItemIndex];
  const { error, generating } = chatItem;
  const errorObject = _.get(error, 'BaseError') || error || {};
  const [open, setOpen] = useState(false);
  const renderFooter = () => {
    if (!generating && !readOnly) {
      return <PanelFooter className="dip-mt-8" chatItemIndex={chatItemIndex} />;
    }
  };
  return (
    <div>
      <div className="dip-flex">
        <AgentIcon
          avatar_type={agentDetails.avatar_type}
          avatar={agentDetails.avatar}
          size={30}
          name={agentDetails.name}
        />
        <div className="dip-ml-16 dip-flex-item-full-width">
          <div className="dip-flex-align-center">
            系统错误，请稍后重试
            {debug && (
              <DipButton type="text" className="dip-ml-8" icon={<FileSearchOutlined />} onClick={() => setOpen(true)} />
            )}
          </div>
          {chatItemIndex === chatList.length - 1 && (
            <div className="dip-flex-align-center">
              <Popover
                content={
                  <div className="dip-flex-column dip-p-12" style={{ gap: 4 }}>
                    <div className="dip-flex-align-center">
                      <span className="dip-text-color">错误原因：</span>
                      <span className="dip-text-color-45">{_.get(errorObject, 'description')}</span>
                    </div>
                    <div className="dip-flex-align-center">
                      <span className="dip-text-color">错误码：</span>
                      <span className="dip-text-color-45">{_.get(errorObject, 'error_code')}</span>
                    </div>
                    <div className="dip-flex-align-center">
                      <span className="dip-text-color">建议：</span>
                      <span className="dip-text-color-45">{_.get(errorObject, 'solution')}</span>
                    </div>
                  </div>
                }
                placement="bottom"
              >
                <div className="dip-font-12 dip-text-link dip-text-color-primary">查看详情</div>
              </Popover>
              {!_.isEmpty(logQueryAgentDetails) && (
                <Tooltip title="AI排查">
                  <Button
                    className="dip-ml-4"
                    size="small"
                    onClick={() => {
                      const stateData = {
                        inputValue: agentDetails.name,
                        fileList: [],
                      };
                      const encodedState = encodeURIComponent(JSON.stringify(stateData));
                      window.open(
                        `${microWidgetProps.history.getBasePath}/usage?id=${logQueryAgentDetails.id}&version=${logQueryAgentDetails.version}&agentAppType=common&state=${encodedState}`
                      );
                    }}
                    type="text"
                    icon={<DipIcon type="icon-dip-ai" className="dip-font-12" />}
                  />
                </Tooltip>
              )}
            </div>
          )}
        </div>
      </div>
      {renderFooter()}
      <DipModal width="70vw" footer={null} title="错误详情" open={open} onCancel={() => setOpen(false)}>
        <AdMonacoEditor
          value={JSON.stringify(error, null, 2)}
          options={{
            minimap: { enabled: false },
            fontSize: 14,
            tabSize: 2,
            insertSpaces: true,
            readOnly: true,
            scrollbar: {
              alwaysConsumeMouseWheel: false, // 禁用Monaco的默认滚轮事件
            },
            lineNumbersMinChars: 4,
            unicodeHighlight: {
              ambiguousCharacters: false, // 关闭中文符号高亮报警
            },
            scrollBeyondLastLine: false, // 禁止滚动超出最后一行
            wordWrap: 'on', // 自动换行，文本始终适应编辑器宽度
            automaticLayout: true, // 自动布局
          }}
          defaultLanguage="json"
          height="auto"
          maxHeight={500}
        />
      </DipModal>
    </div>
  );
};

export default ErrorPanel;
