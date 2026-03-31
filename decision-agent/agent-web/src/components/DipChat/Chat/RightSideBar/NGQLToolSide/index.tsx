import styles from './index.module.less';
import { Splitter } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import DipButton from '@/components/DipButton';
import React from 'react';
import DipIcon from '@/components/DipIcon';
import { useDipChatStore } from '@/components/DipChat/store';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';
import ADTable from '@/components/ADTable';
import { DipChatItemContentProgressType } from '@/components/DipChat/interface';
import { format } from 'sql-formatter';
const NGQLToolSide = () => {
  const {
    dipChatStore: { activeChatItemIndex, chatList, activeProgressIndex },
    closeSideBar,
  } = useDipChatStore();
  const chatItem = chatList[activeChatItemIndex];
  const activeProgress: DipChatItemContentProgressType = chatItem.content.progress[activeProgressIndex] || {};
  const getSql = () => {
    const sqlStr = activeProgress.ngqlResult?.sql;
    if (sqlStr) {
      return sqlStr;
    }
    return JSON.stringify(activeProgress.originalAnswer || '', null, 2);
  };
  return (
    <Splitter layout="vertical">
      <Splitter.Panel>
        <div className="dip-flex-column dip-p-12 dip-full dip-overflow-hidden">
          <div className="dip-flex-space-between">
            <div className="dip-flex-item-full-width dip-flex-align-center">
              <DipIcon className="dip-font-16" type="icon-dip-color-ngl" />
              <div title={activeProgress.title} className="dip-ml-8 dip-flex-item-full-width dip-ellipsis">
                {activeProgress.title}
              </div>
            </div>
            <DipButton size="small" type="text" onClick={closeSideBar}>
              <CloseOutlined className="dip-text-color-45 dip-font-16" />
            </DipButton>
          </div>
          <div className="dip-flex-item-full-height dip-mt-12">
            <AdMonacoEditor
              className={styles.editor}
              value={getSql()}
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
              defaultLanguage="cypher"
            />
          </div>
        </div>
      </Splitter.Panel>
      <Splitter.Panel>
        <div className="dip-p-12 dip-full dip-flex-column dip-overflow-hidden">
          <div className="dip-mb-16">输出</div>
          <div className="dip-flex-item-full-height">
            <ADTable
              autoScrollY
              showHeader={false}
              columns={activeProgress.ngqlResult?.tableColumns}
              dataSource={activeProgress.ngqlResult?.tableData}
            />
          </div>
        </div>
      </Splitter.Panel>
    </Splitter>
  );
};

export default NGQLToolSide;
