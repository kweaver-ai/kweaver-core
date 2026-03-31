import styles from './index.module.less';
import { Segmented, Splitter } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import DipButton from '@/components/DipButton';
import React, { useEffect, useState } from 'react';
import DipIcon from '@/components/DipIcon';
import { useDipChatStore } from '@/components/DipChat/store';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';
import ADTable from '@/components/ADTable';
import DipEcharts from '@/components/DipEcharts';
import { DipChatItemContentProgressType } from '@/components/DipChat/interface';
import _ from 'lodash';
import { useDeepCompareEffect } from '@/hooks';

const SqlToolSide = () => {
  const {
    dipChatStore: { activeChatItemIndex, chatList, activeProgressIndex },
    closeSideBar,
  } = useDipChatStore();
  const chatItem = chatList[activeChatItemIndex];
  const activeProgress: DipChatItemContentProgressType = chatItem.content.progress[activeProgressIndex] || {};
  const [viewMode, setViewMode] = useState('chart');
  const [chartOptions, setChartOptions] = useState({});
  const [error, setError] = useState('');
  const originalAnswer = activeProgress.originalAnswer;
  useDeepCompareEffect(() => {
    const echartsOptions = activeProgress.chartResult?.echartsOptions || {};
    if (!_.isEmpty(echartsOptions)) {
      setChartOptions(activeProgress.chartResult?.echartsOptions || {});
    } else {
      // setChartOptions(originalAnswer ?? '');
      setError(originalAnswer ?? '');
    }
    return () => {
      setChartOptions({});
      setError('');
    };
  }, [activeProgressIndex, activeProgress.chartResult?.echartsOptions]);
  return (
    <Splitter layout="vertical">
      <Splitter.Panel>
        <div className="dip-flex-column dip-p-12 dip-full dip-overflow-hidden">
          <div className="dip-flex-space-between">
            <div className="dip-flex-item-full-width dip-flex-align-center">
              <DipIcon className="dip-font-16" type="icon-dip-color-echarts" />
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
              value={JSON.stringify(error ? error : chartOptions, null, 2)}
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
          <div className="dip-mb-16 dip-flex-space-between">
            <span>输出</span>
            <Segmented
              size="small"
              value={viewMode}
              onChange={value => {
                setViewMode(value);
              }}
              options={[
                {
                  value: 'chart',
                  icon: <DipIcon type="icon-dip-echats" />,
                },
                {
                  value: 'table',
                  icon: <DipIcon type="icon-dip-table" />,
                },
              ]}
            />
          </div>
          <div className="dip-flex-item-full-height dip-w-100">
            {viewMode === 'chart' ? (
              <DipEcharts options={chartOptions} />
            ) : (
              <ADTable
                autoScrollY
                size="small"
                showHeader={false}
                columns={activeProgress.chartResult?.tableColumns}
                dataSource={activeProgress.chartResult?.tableData}
              />
            )}
          </div>
        </div>
      </Splitter.Panel>
    </Splitter>
  );
};

export default SqlToolSide;
