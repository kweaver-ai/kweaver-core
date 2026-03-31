import AdTab from '@/components/AdTab';
import Overview from './Overview';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';

const ExecutionDetails = ({ step, qualityInsightAgentDetails, dateValue, runId }: any) => {
  return (
    <div className="dip-full dip-flex-column">
      <div style={{ backgroundColor: '#F9FAFC' }} className="dip-pl-16 dip-pr-16 dip-pt-12 dip-pb-12">
        <div style={{ fontWeight: 600 }} className="dip-text-color">
          步骤详情
        </div>
        <div className="dip-mt-8 dip-flex dip-text-color-65 dip-font-12">{step.name}</div>
      </div>
      <AdTab
        items={[
          {
            key: 'overview',
            label: '概览',
            children: (
              <Overview
                dateValue={dateValue}
                runId={runId}
                step={step}
                qualityInsightAgentDetails={qualityInsightAgentDetails}
              />
            ),
          },
          {
            key: 'inputOutput',
            label: '输入/输出',
            children: (
              <AdMonacoEditor
                value={JSON.stringify(step.sourceData, null, 2)}
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
            ),
          },
        ]}
      />
    </div>
  );
};

export default ExecutionDetails;
