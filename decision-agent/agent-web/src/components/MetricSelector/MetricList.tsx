import { useState, useMemo, useRef, useEffect, useCallback } from 'react';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { debounce, throttle } from 'lodash';
import { FixedSizeList } from 'react-window';
import AutoSizer from 'react-virtualized-auto-sizer';
import { Input, message, Spin, Tooltip } from 'antd';
import { CheckOutlined } from '@ant-design/icons';
import { getMetricModels } from '@/apis/data-model';
import Empty from '@/components/Empty';
import { useMetricSelectorStore } from './store';
import { LoadStatus, MetricConstants } from './types';
import styles from './MetricList.module.less';
import LoadFailed from '../LoadFailed';
import DipIcon from '@/components/DipIcon';

const limit = 20;

const MetricList = () => {
  const {
    metricSelectorStore,
    removeSelectedMetric,
    appendSelectedMetric,
    updateAllMetricGroup,
    isAllMetricGroupSetted,
  } = useMetricSelectorStore();

  // 懒加载的offset
  const offsetRef = useRef<number>(0);
  // 是否还有数据未加载完
  const hasMoreRef = useRef<boolean>(true);
  // 是否正在加载更多
  const isLoadingMoreRef = useRef<boolean>(false);

  // 搜索关键字
  const searchKeyRef = useRef<string>('');
  const requestRef = useRef<any>(null);

  // 加载状态
  const [loadStatus, setLoadStatus] = useState(LoadStatus.Empty);
  // 数据
  const [data, setData] = useState<{ id: string; name: string }[]>([]);
  const [searchKey, setSearchKey] = useState<string>('');

  // 搜索
  const fetchMetric = useCallback(async () => {
    if (!metricSelectorStore.selectedGroup) return;
    if (!hasMoreRef.current) return;

    try {
      // 只有第一次加载，才设置loadStatus
      if (!isLoadingMoreRef.current) {
        setLoadStatus(LoadStatus.Loading);
      }

      // 开启新的请求之前，先取消之前的请求
      requestRef.current?.abort?.();

      requestRef.current = getMetricModels({
        name_pattern: searchKeyRef.current,
        offset: offsetRef.current,
        limit,
        ...(metricSelectorStore.selectedGroup.id !== '__all'
          ? {
              group_id: metricSelectorStore.selectedGroup.id as unknown as number,
            }
          : {}),
      });

      const { entries, total_count } = await requestRef.current;

      // 设置offset
      offsetRef.current += entries.length;
      // 设置hasMore
      hasMoreRef.current = entries.length === limit;

      if (isLoadingMoreRef.current) {
        // 加载更多：在末尾添加数据
        setData(prev => [...prev, ...entries]);
      } else {
        setData(entries);
        setLoadStatus(entries.length ? LoadStatus.Normal : LoadStatus.Empty);
      }

      // 更新所有指标分组内指标模型数量
      if (metricSelectorStore.selectedGroup.id === '__all' && !isAllMetricGroupSetted) {
        updateAllMetricGroup({ [MetricConstants.MetricModelCount]: total_count });
      }

      isLoadingMoreRef.current = false;
    } catch (ex: any) {
      // 请求取消，无需处理错误
      if (ex === 'CANCEL') return;

      if (ex?.description) {
        message.error(ex.description);
      }

      // 只有第一次加载，才设置loadStatus
      if (!isLoadingMoreRef.current) {
        setLoadStatus(LoadStatus.Failed);
      }

      isLoadingMoreRef.current = false;
    }
  }, [metricSelectorStore.selectedGroup, isAllMetricGroupSetted, updateAllMetricGroup]);

  // 搜索（防抖）
  const searchMetricDebounce = useMemo(() => debounce(fetchMetric, 300), [fetchMetric]);
  // 滑动（节流）
  const handleScroll = useMemo(() => {
    return throttle(
      (
        { scrollOffset, scrollUpdateWasRequested }: any,
        { listHeight, itemSize, dataLength }: { listHeight: number; itemSize: number; dataLength: number }
      ) => {
        if (scrollUpdateWasRequested || isLoadingMoreRef.current || !hasMoreRef.current) return;

        const totalHeight = itemSize * dataLength;
        const threshold = 100;
        const isNearBottom = scrollOffset + listHeight >= totalHeight - threshold;

        if (isNearBottom) {
          isLoadingMoreRef.current = true;
          fetchMetric();
        }
      },
      150
    );
  }, [fetchMetric]);

  const getRow = ({ index, style, data }: any) => {
    const item = data[index];
    // 是否选中
    const isSelected = metricSelectorStore.selectedMetrics.some(metric => metric.id === item.id);

    return (
      <div
        style={style}
        className={classNames('dip-pointer dip-flex-align-center dip-pl-20 dip-pr-20 dip-gap-10', styles['item'], {
          [styles['selected']]: isSelected,
        })}
        onClick={() => {
          if (isSelected) {
            removeSelectedMetric(item);
          } else {
            appendSelectedMetric(item);
          }
        }}
      >
        <DipIcon type={'icon-dip-color-zhibiaometirc'} className="dip-font-24" />
        <span className="dip-flex-column dip-overflow-hidden">
          <Tooltip title={item.name}>
            <span className="dip-ellipsis dip-w-fit-content dip-max-w-100">{item.name}</span>
          </Tooltip>
          <Tooltip title={item.id}>
            <span className="dip-ellipsis dip-text-color-45 dip-w-fit-content dip-max-w-100">{item.id}</span>
          </Tooltip>
        </span>
        {isSelected && <CheckOutlined className={styles['check-icon']} />}
      </div>
    );
  };

  // 当选中的分组id发生变化时，获取该分组下的指标模型列表
  useEffect(() => {
    // 特殊情况：未分组的id是空字符串
    if (metricSelectorStore.selectedGroup && 'id' in metricSelectorStore.selectedGroup) {
      offsetRef.current = 0;
      hasMoreRef.current = true;
      // 切换分组时，清空搜索关键字
      searchKeyRef.current = '';
      setSearchKey('');
      fetchMetric();
    }
  }, [metricSelectorStore.selectedGroup?.id]);

  // 组件卸载时取消防抖
  useEffect(() => {
    return () => searchMetricDebounce.cancel();
  }, [searchMetricDebounce]);

  // 组件卸载时取消 throttle
  useEffect(() => {
    return () => handleScroll.cancel();
  }, [handleScroll]);

  return (
    <div className="dip-w-100 dip-h-100 dip-flex-column">
      <Input
        allowClear
        placeholder={intl.get('dataAgent.searchIndicatorName')}
        className={classNames(styles['input'], 'dip-mt-16 dip-mr-16 dip-mb-8')}
        value={searchKey}
        // 未选中分组时，输入框禁用
        disabled={!metricSelectorStore.selectedGroup}
        onChange={e => {
          const value = e.target.value;
          offsetRef.current = 0;
          hasMoreRef.current = true;
          searchKeyRef.current = value;
          setSearchKey(value);
          searchMetricDebounce();
        }}
      />
      <div className="dip-flex-item-full-height">
        <AutoSizer>
          {({ width, height }) => {
            return loadStatus === LoadStatus.Normal ? (
              <FixedSizeList
                height={height}
                width={width}
                itemCount={data.length}
                itemData={data}
                itemSize={49}
                onScroll={params => {
                  handleScroll(params, { listHeight: height, itemSize: 49, dataLength: data.length });
                }}
              >
                {getRow}
              </FixedSizeList>
            ) : loadStatus === LoadStatus.Empty ? (
              <Empty
                className="dip-h-100 dip-flex-column-center"
                style={{ width, height }}
                description={searchKey ? intl.get('dataAgent.searchResultIsEmpty') : intl.get('dataAgent.noData')}
              />
            ) : loadStatus === LoadStatus.Loading ? (
              <div style={{ width, height }} className="dip-flex-column-center">
                <Spin />
              </div>
            ) : (
              <LoadFailed
                className="dip-h-100 dip-flex-column-center"
                style={{ width, height }}
                onRetry={() => {
                  offsetRef.current = 0;
                  hasMoreRef.current = true;
                  searchMetricDebounce();
                }}
              />
            );
          }}
        </AutoSizer>
      </div>
    </div>
  );
};

export default MetricList;
