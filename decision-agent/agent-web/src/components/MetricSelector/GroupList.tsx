import { useEffect, useState, useRef, useMemo, useCallback } from 'react';
import intl from 'react-intl-universal';
import { throttle } from 'lodash';
import classNames from 'classnames';
import { FixedSizeList } from 'react-window';
import AutoSizer from 'react-virtualized-auto-sizer';
import { message, Spin, Tooltip } from 'antd';
import { getMetricModalGroups, type MetricModalGroupType } from '@/apis/data-model';
import LoadFailed from '@/components/LoadFailed';
import { useMetricSelectorStore } from './store';
import { LoadStatus, MetricConstants } from './types';
import styles from './GroupList.module.less';

// 每一页获取的数量
const limit = 20;

const GroupList = () => {
  const { metricSelectorStore, setSelectedGroup } = useMetricSelectorStore();

  // 懒加载的offset
  const offsetRef = useRef<number>(0);
  // 是否还有数据未加载完
  const hasMoreRef = useRef<boolean>(true);
  // 是否正在加载更多
  const isLoadingMoreRef = useRef<boolean>(false);

  // 加载数据的状态
  const [loadStatus, setLoadStatus] = useState<LoadStatus>(LoadStatus.Loading);
  // 分组数据
  const [data, setData] = useState<MetricModalGroupType[]>([]);
  // 界面上显示的数据列表，第一位永远是【所有指标模型】
  const displayData = useMemo(
    () => [metricSelectorStore.allMetricGroup, ...data],
    [metricSelectorStore.allMetricGroup, data]
  );

  // 获取分组数据
  const fetchGroups = useCallback(async () => {
    if (!hasMoreRef.current) return;

    try {
      // 只有第一次加载，才设置loadStatus
      if (!isLoadingMoreRef.current) {
        setLoadStatus(LoadStatus.Loading);
      }

      const { entries } = await getMetricModalGroups({
        offset: offsetRef.current,
        limit,
      });
      // 设置offset
      offsetRef.current += entries.length;
      // 设置hasMore
      hasMoreRef.current = entries.length === limit;

      if (isLoadingMoreRef.current) {
        // 加载更多：在末尾添加数据
        setData(prev => [...prev, ...entries]);
      } else {
        setData(entries);
        setLoadStatus(LoadStatus.Normal);
        // 默认选中全部指标组
        setSelectedGroup(metricSelectorStore.allMetricGroup);
      }
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }

      // 只有第一次加载的时候，才设置loadStatus
      if (!isLoadingMoreRef.current) {
        setLoadStatus(LoadStatus.Failed);
      }
    } finally {
      isLoadingMoreRef.current = false;
    }
  }, [metricSelectorStore.allMetricGroup, setSelectedGroup]);

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
          fetchGroups();
        }
      },
      150
    );
  }, [fetchGroups]);

  const getRow = ({ index, style, data }: any) => {
    const item = data[index];
    const name = item.name || intl.get('dataAgent.ungrouped');

    return (
      <div
        style={style}
        className={classNames('dip-pointer dip-flex-align-center dip-pl-20', styles['item'], {
          [styles['selected']]: metricSelectorStore.selectedGroup?.id === item.id,
        })}
        onClick={() => {
          setSelectedGroup(item);
        }}
      >
        <Tooltip title={name}>
          <span className="dip-ellipsis">{name}</span>
        </Tooltip>
        {MetricConstants.MetricModelCount in item && (
          <span className="dip-pr-10 dip-pre-whitespace">{` (${item.metric_model_count})`}</span>
        )}
      </div>
    );
  };

  useEffect(() => {
    fetchGroups();
  }, []);

  // 组件卸载时取消 throttle
  useEffect(() => {
    return () => handleScroll.cancel();
  }, [handleScroll]);

  return (
    <div className="dip-flex-column dip-h-100">
      <div className={classNames(styles['header'], 'dip-pl-10')}>{intl.get('dataAgent.selectGroup')}</div>

      <div className="dip-flex-item-full-height">
        <AutoSizer>
          {({ width, height }) => {
            return loadStatus === LoadStatus.Normal ? (
              <FixedSizeList
                height={height}
                width={width}
                itemCount={displayData.length}
                itemData={displayData}
                itemSize={36}
                onScroll={params => {
                  handleScroll(params, { listHeight: height, itemSize: 36, dataLength: displayData.length });
                }}
              >
                {getRow}
              </FixedSizeList>
            ) : loadStatus === LoadStatus.Failed ? (
              <LoadFailed
                className="dip-h-100 dip-flex-column-center"
                style={{ width, height }}
                onRetry={fetchGroups}
              />
            ) : (
              <div style={{ width, height }} className="dip-flex-column-center">
                <Spin />
              </div>
            );
          }}
        </AutoSizer>
      </div>
    </div>
  );
};

export default GroupList;
