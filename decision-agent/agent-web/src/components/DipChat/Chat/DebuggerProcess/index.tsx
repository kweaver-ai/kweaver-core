import { useRef, useState } from 'react';
import _ from 'lodash';
import intl from 'react-intl-universal';
import EmptyIcon from '@/assets/icons/empty.svg';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';
import { getBlockIcon } from './utils';
import './style.less';
import { useDipChatStore } from '@/components/DipChat/store';
import useDeepCompareMemo from '@/hooks/useDeepCompareMemo';
import ScrollBarContainer from '@/components/ScrollBarContainer';
import classNames from 'classnames';
import ResizeObserver from '@/components/ResizeObserver';
import { Tooltip } from 'antd';

const DebuggerProcess = () => {
  const {
    dipChatStore: { chatList, singleStreamResult },
  } = useDipChatStore();
  const chatItem = chatList[chatList.length - 1];
  const sourceData = chatItem?.error ? (singleStreamResult[0] ?? {}) : chatItem?.sourceData || {};
  const [activeIndex, setActiveIndex] = useState(-1);
  const autoUpdateActiveIndex = useRef(true);
  const progressWrapperRef = useRef<HTMLDivElement>(null);

  const progress = useDeepCompareMemo(() => {
    const operatorNames: any = {
      llm: '大模型输出',
      skill: '技能/工具调用',
      assign: '赋值操作',
    };
    const progressRes = (_.get(sourceData, 'message.content.middle_answer.progress', []) || []).filter(
      (item: any) => !!item
    );
    const tempArr = [
      ...progressRes.map((item: any, index: number) => ({
        // name: item?.skill_info?.name || 'LLM',
        name:
          item.stage === 'assign'
            ? operatorNames[item.stage]
            : `${item.agent_name || 'LLM'} (${operatorNames[item.stage]})`,
        key: index,
        icon: getBlockIcon(item.stage),
        content: { ...item },
      })),
    ];
    if (autoUpdateActiveIndex.current) {
      setActiveIndex(tempArr.length - 1);
    }
    return tempArr;
  }, [sourceData]);

  const getMonacoValue = () => {
    if (activeIndex > -1 && progress.length > activeIndex) {
      const item = progress[activeIndex];
      return JSON.stringify(item.content, null, 2);
    }
    return '';
  };
  const getActiveTitle = () => {
    if (activeIndex > -1 && progress.length > activeIndex) {
      const item = progress[activeIndex];
      return item.name;
    }
    return '';
  };

  return _.isEmpty(progress) ? (
    <div className="dip-column-center dip-full" style={{ marginTop: -100 }}>
      <EmptyIcon />
      <p style={{ color: '#779EEA' }}>{intl.get('agent.viewDebuggerTip')}</p>
    </div>
  ) : (
    <div className="dip-full debugRoot dip-flex dip-position-r">
      <div className="debugRoot-title dip-text-color-primary">{getActiveTitle()}</div>
      <div className="debugRoot-progress dip-h-100 dip-border-r">
        <ScrollBarContainer className="dip-h-100" ref={progressWrapperRef}>
          <ResizeObserver
            onResize={() => {
              if (autoUpdateActiveIndex.current) {
                progressWrapperRef.current!.scrollTop = progressWrapperRef.current!.scrollHeight;
              }
            }}
          >
            <div>
              {progress.map((item: any, index: number) => {
                const { end_time, start_time, token_usage, status, stage } = item.content || {};
                return (
                  <div
                    className={classNames('debugRoot-progress-item', {
                      'debugRoot-progress-item-active': activeIndex === index,
                    })}
                    key={item.key}
                    onClick={() => {
                      autoUpdateActiveIndex.current = false;
                      setActiveIndex(index);
                    }}
                  >
                    <Tooltip title={item.name}>
                      <div className="dip-ellipsis">{item.name}</div>
                    </Tooltip>
                    {status === 'completed' && (
                      <>
                        <div className="dip-flex-space-between dip-text-color-45 dip-font-12">
                          <span>耗时:</span>
                          <span>{(end_time - start_time).toFixed(2)}s</span>
                        </div>
                        {stage === 'llm' && (
                          <>
                            <div className="dip-flex-space-between dip-text-color-45 dip-font-12 dip-mt-4">
                              <span>prompt_cached_tokens:</span>
                              <span>{_.get(token_usage, 'prompt_tokens_details.cached_tokens', 0)}</span>
                            </div>
                            <div className="dip-flex-space-between dip-text-color-45 dip-font-12 dip-mt-4">
                              <span>tokens:</span>
                              <span>{_.get(token_usage, 'total_tokens', 0)}</span>
                            </div>
                          </>
                        )}
                      </>
                    )}
                  </div>
                );
              })}
            </div>
          </ResizeObserver>
        </ScrollBarContainer>
      </div>
      <div className="dip-flex-item-full-width dip-h-100">
        <AdMonacoEditor
          // height="auto"
          // maxHeight={monacoEditorMaxHeight}
          value={getMonacoValue()}
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
        />
      </div>
    </div>
  );
};

export default DebuggerProcess;
