import { useEffect, useState } from 'react';
import _ from 'lodash';
import { useLatestState } from '@/hooks';
import { getChatUrl } from '@/apis/super-assistant';
import { isJSONString } from '@/utils/handle-function';
import TraceAnalysisTitle from '../Title';
import { Button, Tag } from 'antd';
import { InfoCircleOutlined, RedoOutlined } from '@ant-design/icons';
import { AdMonacoEditor } from '@/components/Editor/AdMonacoEditor';
import DipModal from '@/components/DipModal';
import DipIcon from '@/components/DipIcon';
import { post } from '@/utils/http';
import styles from './index.module.less';
import FailureIcon from '@/assets/images/Failure.svg';
import AiRobot from './AiRobot';
import Loading from './Loading';
import Title from '../Title';

export type OptimizationSuggestionProps = {
  qualityInsightAgentDetails: any;
  analysisLevel: 'agent' | 'session' | 'run';
  startTime: number;
  endTime: number;
  id: string; // Agent ID 、Session ID、Run ID
};

const OptimizationSuggestion = ({
  qualityInsightAgentDetails,
  analysisLevel,
  startTime,
  endTime,
  id,
}: OptimizationSuggestionProps) => {
  // const mounted = useRef(false);
  // const [response, send, stop] = useStreamingOut();
  const [streamGenerating, setStreamGenerating, getStreamGenerating] = useLatestState(false);
  const [data, setData] = useState<any>({});
  const [error, setError] = useState<any>({});
  const [open, setOpen] = useState(false);

  useEffect(() => {
    return () => {
      // stop();
    };
  }, []);

  // useDeepCompareEffect(() => {
  //   if (!mounted.current) {
  //     mounted.current = true;
  //     return;
  //   }
  //   if (response.error) {
  //     setError(response.error);
  //   } else {
  //     const contentObj = isJSONString(response.content) ? JSON.parse(response.content) : {};
  //     const { message, error } = contentObj;
  //     if (!_.isEmpty(error)) {
  //       setError(error);
  //     } else if (message && _.isObject(message)) {
  //       const text = _.get(message, 'content.final_answer.answer.text', '');
  //       console.log(text, 'text');
  //       if (isJSONString(text)) {
  //         setData(JSON.parse(text));
  //       }
  //     }
  //   }
  //   if (!response.generating) {
  //     setStreamGenerating(false);
  //   }
  // }, [response]);

  const sendChat = async () => {
    if (qualityInsightAgentDetails.id && !getStreamGenerating()) {
      setStreamGenerating(true);
      try {
        const res = await post(getChatUrl(qualityInsightAgentDetails.id, false, false), {
          body: {
            agent_id: qualityInsightAgentDetails.id,
            agent_version: qualityInsightAgentDetails.version,
            stream: false,
            inc_stream: false,
            executor_version: 'v2',
            custom_querys: {
              analysis_level: analysisLevel,
              id,
              start_time: startTime,
              end_time: endTime,
            },
          },
        });
        setStreamGenerating(false);
        if (res) {
          const { message, error } = res;
          if (!_.isEmpty(error)) {
            setError(error);
          } else if (message && _.isObject(message)) {
            let text = _.get(message, 'content.final_answer.answer.text', '');
            text = text.replace('```json', '');
            text = text.replace('```', '');
            if (isJSONString(text)) {
              setData(JSON.parse(text));
            }
          }
        }
      } catch (error) {
        setStreamGenerating(false);
        console.log(error, 'error');
        setError(error);
      }
    }
  };

  const getLevel = (level: 'critical' | 'high' | 'medium' | 'low') => {
    if (level === 'critical') {
      return <Tag color="red">严重</Tag>;
    }
    if (level === 'high') {
      return <Tag color="error">高优先级</Tag>;
    }
    if (level === 'medium') {
      return <Tag color="warning">中优先级</Tag>;
    }
    if (level === 'low') {
      return <Tag color="processing">低优先级</Tag>;
    }
  };

  const getColor = (level: 'critical' | 'high' | 'medium' | 'low') => {
    if (level === 'critical') {
      return {
        color: '#cf1322',
        bgColor: '#fff1f0',
      };
    }
    if (level === 'high') {
      return {
        color: '#ff4d4f',
        bgColor: '#fff2f0',
      };
    }
    if (level === 'medium') {
      return {
        color: '#faad14',
        bgColor: '#fffbe6',
      };
    }
    if (level === 'low') {
      return {
        color: '#1677ff',
        bgColor: '#e6f4ff',
      };
    }
  };

  const renderCategory = (category: 'performance' | 'efficiency' | 'stability' | 'quality') => {
    if (category === 'performance') {
      return <Tag color="default">性能</Tag>;
    }
    if (category === 'efficiency') {
      return <Tag color="default">效率</Tag>;
    }
    if (category === 'stability') {
      return <Tag color="default">稳定性</Tag>;
    }
    if (category === 'quality') {
      return <Tag color="default">质量</Tag>;
    }
  };

  const renderContent = () => {
    if (!_.isEmpty(error)) {
      return (
        <div className="dip-flex-column-center dip-p-24">
          <FailureIcon />
          <span>AI分析失败</span>
          <div className="dip-text-color-65 dip-mt-8">
            抱歉，Al智能分析过程中出现了错误。这可能是由于网络连接问题或服务暂时不可用导致的。
          </div>
          <div className="dip-mt-20">
            <Button
              type="primary"
              icon={<RedoOutlined />}
              onClick={() => {
                setError({});
                setData({});
                sendChat();
              }}
            >
              重新分析
            </Button>
            <Button className="dip-ml-8" icon={<InfoCircleOutlined />} onClick={() => setOpen(true)}>
              查看详情
            </Button>
          </div>
        </div>
      );
    }
    if (streamGenerating) {
      return (
        <div className="dip-flex-column-center" style={{ height: 200 }}>
          <Loading />
          <div className="dip-mt-16">AI分析中...</div>
        </div>
      );
    }
    if (_.isEmpty(data)) {
      return (
        <div className="dip-flex-column-center dip-mt-24 dip-mb-24">
          <AiRobot />
          <div className="dip-text-color-65 dip-mt-8">点击"AI智能分析"按钮，让AI为您的智能体生成专业的优化建议</div>
          <div className="dip-mt-20">
            <Button icon={<DipIcon type="icon-dip-color-ai-analyse" />} loading={streamGenerating} onClick={sendChat}>
              AI智能分析
            </Button>
          </div>
        </div>
      );
    }
    return (
      <div className="dip-mt-16">
        <div className="dip-text-color dip-mt-16">总结：{data.summary}</div>
        {data.findings?.map((item: any) => {
          const { bgColor, color } = getColor(item.severity) || {};
          return (
            <div key={item.issue_id} className="dip-mt-12">
              <div className={styles.card}>
                <div className="dip-flex-column">
                  <div>
                    <span>{getLevel(item.severity)}</span>
                    <span>{renderCategory(item.category)}</span>
                  </div>

                  <div className="dip-mt-16 dip-flex-column">
                    <Title color={color} title={<span className="dip-font-14">问题：</span>} />
                    <span className="dip-mt-12">{item.description}</span>
                  </div>
                  <div className="dip-mt-16 dip-flex-column">
                    <Title color={color} title={<span className="dip-font-14">依据：</span>} />
                    <div className="dip-flex-column dip-mt-12" style={{ gap: 4 }}>
                      {item.evidence?.map((item: any) => (
                        <div key={item} className="dip-flex-align-center">
                          <div style={{ width: 4, height: 4, background: '#677489', borderRadius: '50%' }} />
                          <span className="dip-ml-8">{item}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                  <div className="dip-mt-16 dip-flex-column">
                    <Title color={color} title={<span className="dip-font-14">影响：</span>} />
                    <div className="dip-mt-12">{item.impact}</div>
                  </div>
                  <div className="dip-mt-16 dip-flex-column">
                    <Title color={color} title={<span className="dip-font-14">建议：</span>} />
                    <div>
                      {item.recommendations?.map((item: any) => (
                        <div
                          className="dip-mt-12 dip-p-12"
                          key={item.action}
                          // style={{ background: getColor(item.severity)?.bgColor, borderRadius: 12 }}
                          style={{ background: bgColor, borderRadius: 12 }}
                        >
                          <div className="dip-text-color">{item.action}</div>
                          <div className="dip-mt-8">{item.details}</div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div>
      <TraceAnalysisTitle
        title={
          <div className="dip-flex-align-center">
            <span>优化建议</span>
            {!_.isEmpty(data) && (
              <span className="dip-flex-align-center dip-ml-12">
                <DipIcon type="icon-dip-color-ai-analyse" />
                <span className="dip-ml-8 dip-font-12 dip-text-color-45">来自AI智能分析</span>
              </span>
            )}
          </div>
        }
      />
      {renderContent()}
      <DipModal width="70vw" footer={null} title="异常详情" open={open} onCancel={() => setOpen(false)}>
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

export default OptimizationSuggestion;
