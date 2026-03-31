import styles from './index.module.less';
import { Splitter } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import DipButton from '@/components/DipButton';
import { useDipChatStore } from '@/components/DipChat/store';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';
import { DipChatItemContentProgressType } from '@/components/DipChat/interface';
import ToolIcon from '@/assets/icons/toolIcon.svg';
import AgentImg from '@/assets/icons/agent3.svg';
import MCPIcon from '@/assets/icons/mcp.svg';
import DipIcon from '@/components/DipIcon';
import intl from 'react-intl-universal';
import { isJSONString } from '@/utils/handle-function';
import Markdown from '@/components/Markdown';
import { removeInvalidCodeBlocks } from '@/components/Markdown/utils';
import ScrollBarContainer from '@/components/ScrollBarContainer';
const SqlToolSide = () => {
  const {
    dipChatStore: { activeChatItemIndex, chatList, activeProgressIndex },
    closeSideBar,
  } = useDipChatStore();
  const chatItem = chatList[activeChatItemIndex];
  const activeProgress: DipChatItemContentProgressType = chatItem.content.progress[activeProgressIndex] || {};
  const skillInfo = activeProgress.skillInfo;
  const renderSkillIcon = () => {
    if (skillInfo.name === 'graph_qa') {
      return <DipIcon type="icon-dip-color-graph" className="dip-font-16" />;
    }
    if (skillInfo.type === 'TOOL') {
      return <ToolIcon style={{ width: '16px', height: '16px' }} />;
    }
    if (skillInfo.type === 'AGENT') {
      return <AgentImg style={{ width: '16px', height: '16px' }} />;
    }
    if (skillInfo.type === 'MCP') {
      return <MCPIcon style={{ width: '16px', height: '16px' }} />;
    }
  };

  const renderOutput = () => {
    const outputStr = activeProgress.commonToolResult?.output;
    if (outputStr && isJSONString(outputStr)) {
      const outputObj = JSON.parse(outputStr);
      if (outputObj.answer && typeof outputObj.answer === 'string') {
        // console.log(outputObj.answer, '我是普通工具的输出啊啊啊');
        return (
          <ScrollBarContainer className="dip-full">
            <Markdown readOnly value={removeInvalidCodeBlocks(outputObj.answer, true)} />
          </ScrollBarContainer>
        );
      }
    }
    return (
      <AdMonacoEditor
        className={styles.editor}
        value={activeProgress.commonToolResult?.output}
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
      />
    );
  };

  return (
    <Splitter layout="vertical">
      <Splitter.Panel>
        <div className="dip-flex-column dip-p-12 dip-full dip-overflow-hidden">
          <div className="dip-flex-space-between">
            <div className="dip-flex-item-full-width dip-flex-align-center">
              {renderSkillIcon()}
              <div title={activeProgress.title} className="dip-ml-8 dip-flex-item-full-width dip-ellipsis">
                {activeProgress.title}
              </div>
            </div>
            <DipButton size="small" type="text" onClick={closeSideBar}>
              <CloseOutlined className="dip-text-color-45 dip-font-16" />
            </DipButton>
          </div>
          <div className="dip-flex-item-full-height dip-mt-8">
            <AdMonacoEditor
              className={styles.editor}
              value={activeProgress.commonToolResult?.input}
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
            />
          </div>
        </div>
      </Splitter.Panel>
      <Splitter.Panel>
        <div className="dip-p-12 dip-full dip-flex-column dip-overflow-hidden">
          <div className="dip-mb-8">{intl.get('dipChat.output')}</div>
          <div className="dip-flex-item-full-height">{renderOutput()}</div>
        </div>
      </Splitter.Panel>
    </Splitter>
  );
};

export default SqlToolSide;
