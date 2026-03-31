import { useMemo, useRef, useEffect } from 'react';
import type { EChartsOption, ECharts } from 'echarts';
import DipEcharts from '@/components/DipEcharts';
import styles from './index.module.less';
import { useDeepCompareMemo, useLatestState } from '@/hooks';

export type ExecutionStep = {
  name: string;
  startTime: number; // 开始时间（毫秒）
  endTime: number; // 结束时间（毫秒）
  status: 'completed' | 'failed'; // 执行状态
  sourceData?: any;
  color: string;
  textPosition: string;
};

export type ExecutionTimelineProps = {
  onStepClick?: (step: ExecutionStep) => void;
  runDetails: any;
};

const ExecutionTimeline = ({ onStepClick, runDetails }: ExecutionTimelineProps) => {
  const chartRef = useRef<{ getEchartsInstance: () => ECharts }>(null);
  const [selectedStepIndex, setSelectedStepIndex, getSelectedStepIndex] = useLatestState<number>(-1);

  const steps: ExecutionStep[] = useDeepCompareMemo(() => {
    const operatorNames: any = {
      llm: '大模型输出',
      skill: '技能/工具调用',
      assign: '赋值操作',
    };
    const arr = runDetails.progress.map((item: any) => ({
      name:
        item.stage === 'assign'
          ? operatorNames[item.stage]
          : `${item.agent_name || 'LLM'} (${operatorNames[item.stage]})`,
      startTime: item.start_time * 1000,
      endTime: item.end_time * 1000,
      status: item.status,
      sourceData: { ...item },
      color: '#677489',
      textPosition: 'insideLeft',
    }));
    if (arr.length > 0) {
      onStepClick?.(arr[0]);
    }
    arr.unshift({
      name: '总耗时',
      startTime: arr[0].startTime,
      endTime: arr[arr.length - 1].endTime,
      status: 'completed',
      color: '#fff',
      textPosition: 'inside',
    });
    return arr;
  }, [runDetails]);

  const minStartTime = useMemo(() => {
    if (!steps.length) {
      return 0;
    }
    return Math.min(...steps.map(step => step.startTime));
  }, [steps]);

  const totalDuration = useMemo(() => {
    if (!steps.length) {
      return 0;
    }
    const maxEnd = Math.max(...steps.map(step => step.endTime));
    return maxEnd - minStartTime;
  }, [steps, minStartTime]);

  const chartOptions = useMemo(() => {
    if (steps.length === 0) {
      return {};
    }

    const maxTimeSeconds = Math.max(Math.ceil(totalDuration / 1000), 1);

    // 准备甘特图数据
    const categories = steps.map(step => step.name);
    const seriesData = steps.map((step, index) => {
      const start = (step.startTime - minStartTime) / 1000; // 转换为秒
      const end = (step.endTime - minStartTime) / 1000;
      const duration = Math.max(step.endTime - step.startTime, 0);
      const durationSeconds = duration / 1000;
      let color;
      if (!step.sourceData) {
        color = '#3DBCF3';
      } else {
        color = step.status === 'completed' ? 'rgb(82, 196, 26)' : '#DE7670';
      }
      const isSelected = selectedStepIndex === index;

      return {
        value: [index, start, end, durationSeconds],
        itemStyle: {
          color: color,
          borderColor: isSelected ? '#126ee3' : 'transparent',
          borderWidth: isSelected ? 2 : 0,
        },
      };
    });

    return {
      // tooltip: {
      //   trigger: 'axis',
      //   axisPointer: {
      //     type: 'shadow',
      //   },
      //   formatter: (params: any) => {
      //     const param = Array.isArray(params) ? params[0] : params;
      //     const stepIndex = param.value[0];
      //     const step = steps[stepIndex];
      //     // console.log(step, '哈哈哈');
      //     if (!step) return '';
      //     const duration = Math.max(step.endTime - step.startTime, 0);
      //     return `
      //       <div style="padding: 8px;">
      //         <div style="font-weight: bold; margin-bottom: 4px;">${step.name}</div>
      //         <div>开始时间: ${dayjs(step.startTime).format('HH:mm:ss')}</div>
      //         <div>结束时间: ${dayjs(step.endTime).format('HH:mm:ss')}</div>
      //         <div>耗时: ${(duration / 1000).toFixed(2)}s</div>
      //         <div>状态: ${step.status === 'completed' ? '成功' : '失败'}</div>
      //       </div>
      //     `;
      //   },
      // },
      grid: {
        left: '2%',
        right: '2%',
        top: '7%',
        bottom: '1%',
      },
      xAxis: {
        type: 'value',
        // name: '时间 (秒)',
        position: 'top',
        nameLocation: 'middle',
        nameGap: 30,
        min: 0,
        max: maxTimeSeconds,
        axisLabel: {
          formatter: (value: number) => `${value}s`,
          margin: 12,
          color: '#677489',
        },
        axisLine: {
          lineStyle: {
            color: '#d9d9d9',
          },
        },
        axisTick: {
          show: false,
        },
        splitLine: {
          show: false,
          lineStyle: {
            type: 'dashed',
            color: '#e8e8e8',
          },
        },
      },
      yAxis: {
        type: 'category',
        data: categories,
        inverse: true,
        axisLabel: {
          show: false,
          margin: 12,
          color: '#000',
        },
        axisLine: {
          show: false,
        },
        axisTick: {
          show: false,
        },
      },
      series: [
        {
          type: 'custom',
          renderItem: (_params: any, api: any) => {
            const categoryIndex = api.value(0);
            const start = api.coord([api.value(1), categoryIndex]);
            const end = api.coord([api.value(2), categoryIndex]);
            const height = api.size([0, 1])[1] * 0.6;
            const itemStyle = api.style();
            const durationSeconds = api.value(3) ?? 0;
            const rowStart = api.coord([0, categoryIndex]);
            const rowEnd = api.coord([maxTimeSeconds, categoryIndex]);
            const rectShape = {
              x: start[0],
              y: start[1] - height / 2,
              width: end[0] - start[0],
              height: height,
            };
            return {
              type: 'group',
              children: [
                {
                  type: 'rect',
                  silent: true,
                  shape: {
                    x: rowStart[0],
                    y: rowStart[1] - height / 2,
                    width: rowEnd[0] - rowStart[0],
                    height: height,
                  },
                  style: {
                    fill: '#F9FAFC',
                  },
                },
                {
                  type: 'rect',
                  shape: rectShape,
                  style: {
                    fill: itemStyle.fill,
                    stroke: itemStyle.borderColor,
                    lineWidth: itemStyle.borderWidth,
                    borderRadius: 4,
                  },
                  textContent: {
                    type: 'text',
                    style: {
                      text: `${steps[categoryIndex].name} ${durationSeconds.toFixed(3)}s`,
                      fill: steps[categoryIndex].color,
                      fontSize: 12,
                    },
                  },
                  textConfig: {
                    position: steps[categoryIndex].textPosition,
                  },
                },
              ],
            };
          },
          data: seriesData,
        },
      ],
    };
  }, [steps, selectedStepIndex, totalDuration, minStartTime]);

  // 绑定点击事件
  useEffect(() => {
    if (!chartRef.current || !onStepClick) return;

    const chartInstance = chartRef.current.getEchartsInstance();
    const handleClick = (params: any) => {
      if (params.data && params.data.value) {
        const stepIndex = params.data.value[0];
        if (stepIndex > 0 && stepIndex !== getSelectedStepIndex()) {
          setSelectedStepIndex(stepIndex);
          const step: any = steps[stepIndex];
          onStepClick(step);
        }
      }
    };

    chartInstance.on('click', handleClick);
    return () => {
      chartInstance.off('click', handleClick);
    };
  }, [steps]);

  return (
    <div className="dip-full dip-flex-column">
      <div className={styles.timelineHeader}>
        <div style={{ fontWeight: 600 }} className="dip-text-color">
          执行时间线
        </div>
        <div className="dip-mt-8 dip-flex dip-text-color-65 dip-font-12">
          <span>
            总耗时: <span>{(runDetails.total_time / 1000).toFixed(3)}s</span>
          </span>
          <span className="dip-ml-12 dip-mr-12">.</span>
          <span>
            <span>{runDetails.progress.length}个</span>步骤
          </span>
          <span className="dip-ml-12 dip-mr-12">.</span>
          <span>
            <span>{runDetails.tool_call_failed_count}个</span>失败
          </span>
        </div>
      </div>
      <div className="dip-flex-item-full-height">
        <DipEcharts ref={chartRef} options={chartOptions as EChartsOption} />
      </div>
    </div>
  );
};

export default ExecutionTimeline;
