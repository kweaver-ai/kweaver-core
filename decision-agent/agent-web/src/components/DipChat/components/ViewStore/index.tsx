import React, { type CSSProperties, useState } from 'react';
import { Button } from 'antd';
import { useDipChatStore } from '../../store';
// import { copyToBoard } from '@/utils/handle-function';
import { useMicroWidgetProps } from '@/hooks';
import DipModal from '@/components/DipModal';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';
import AdTab from '@/components/AdTab';
const ViewStore: React.FC = () => {
  const microWidgetProps = useMicroWidgetProps();
  const { dipChatStore } = useDipChatStore();
  const isDev = process.env.NODE_ENV === 'development';
  const [open, setOpen] = useState(false);
  const [tabValue, setTabValue] = useState('last');

  let style: CSSProperties = {
    opacity: 0,
    userSelect: 'none',
    pointerEvents: 'none',
  };
  if (isDev) {
    style = {};
  }

  return (
    <div className="dip-position-a" style={{ right: 0, top: 0, zIndex: 9999, ...style }}>
      <Button
        id="dip-chat-debug-btn"
        size="small"
        type="primary"
        onClick={() => {
          if (!isDev) {
            setOpen(true);
          } else {
            console.log(dipChatStore, 'dipChatStore');
            // const aaa = JSON.stringify(dipChatStore);
            // copyToBoard(aaa);
            console.log(microWidgetProps, '主应用传过来的属性');
          }
        }}
      >
        查看dipChatStore
      </Button>

      {open && dipChatStore.singleStreamResult.length > 0 && (
        <DipModal footer={null} title="流式结果" fullScreen open onCancel={() => setOpen(false)}>
          <div className="dip-full dip-flex-column">
            <AdTab
              activeKey={tabValue}
              onChange={value => {
                setTabValue(value);
              }}
              items={[
                {
                  label: '最后一次输出流式结果',
                  key: 'last',
                  children: (
                    <AdMonacoEditor
                      value={JSON.stringify(
                        dipChatStore.singleStreamResult[dipChatStore.singleStreamResult.length - 1],
                        null,
                        2
                      )}
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
                      // onMount={editor => {
                      //   editor.getAction('editor.foldAll').run();
                      // }}
                    />
                  ),
                },
              ]}
            />
          </div>
        </DipModal>
      )}
    </div>
  );
};

export default ViewStore;
