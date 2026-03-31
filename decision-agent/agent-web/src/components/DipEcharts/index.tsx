import React, { CSSProperties, forwardRef, useEffect, useImperativeHandle, useRef } from 'react';
import * as echarts from 'echarts';
import type { EChartsOption, ECharts } from 'echarts';
import classNames from 'classnames';
import { useDeepCompareEffect, useLatestState } from '@/hooks';
import ResizeObserver from '@/components/ResizeObserver';

export type DipEchartsProps = {
  className?: string;
  style?: CSSProperties;
  options: EChartsOption;
};

export type DipEchartsRef = {
  getEchartsInstance: () => ECharts;
};
const DipEcharts = forwardRef<DipEchartsRef, DipEchartsProps>((props, ref) => {
  const { className, style, options } = props;
  const chartsInstance = useRef<ECharts>();
  const chartsWrapper = useRef<HTMLDivElement | null>(null);
  const [rendered, setRendered, getRendered] = useLatestState(false);

  useImperativeHandle(ref, () => ({
    getEchartsInstance,
  }));

  const getEchartsInstance = () => chartsInstance.current!;

  useEffect(() => {
    chartsInstance.current = echarts.init(chartsWrapper.current);
    chartsInstance.current.on('finished', handleFinished);
    return () => {
      chartsInstance.current?.dispose();
      chartsInstance.current?.off('finished', handleFinished);
    };
  }, []);

  useDeepCompareEffect(() => {
    chartsInstance.current?.setOption(options);
  }, [options]);

  const handleFinished = () => {
    if (!getRendered()) {
      setRendered(true);
    }
  };

  return (
    <ResizeObserver
      onResize={() => {
        if (getRendered()) {
          chartsInstance.current?.resize();
        }
      }}
    >
      <div className="dip-full">
        <div ref={chartsWrapper} style={style} className={classNames('dip-echarts dip-full', className)} />
      </div>
    </ResizeObserver>
  );
});

export default DipEcharts;
