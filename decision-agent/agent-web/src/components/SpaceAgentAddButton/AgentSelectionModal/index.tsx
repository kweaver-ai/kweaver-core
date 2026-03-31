import { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { debounce, uniq } from 'lodash';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { Modal, Button, Input, message, Spin, Empty } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { VariableSizeGrid } from 'react-window';
import AutoSizer from 'react-virtualized-auto-sizer';
import { getAgentsByPost } from '@/apis/agent-factory';
import LoadFailedIcon from '@/assets/icons/load-failed.svg';
import { useUserAvatars } from '@/hooks/useUserAvatars';
import { formatTimeSlash } from '@/utils/handle-function/FormatTime';
import BaseCard, { rowHeight, loadingMoreRowHeight, gap } from '@/components/BaseCard';
import styles from './index.module.less';

const pageSize = 120; // 每页展示数量
enum LoadStatus {
  Loading = 'loading', // 加载中
  Empty = 'empty', // 内容为空
  Normal = 'normal', // 有数据
  Failed = 'failed', // 失败
  LoadingMore = 'loadingMore', // 加载更多
}

interface Props {
  onCancel: () => void;
  onConfirm: (selections: any[]) => void;
}

const columnCount = 2; // 每行显示多少列

const AgentSelectionModal = ({ onCancel, onConfirm }: Props) => {
  const listCountRef = useRef<number>(0);
  const gridRef = useRef(null); // 使用ref来调用Grid的recomputeGridSize方法
  const searchKeyRef = useRef<string>('');
  const loadingMoreRef = useRef<boolean>(false); // 是否正在加载更多
  const nextPaginationMarkerStrRef = useRef<string>('');

  const { userAvatars, addUserIds } = useUserAvatars([], 20); // 头像url数据

  const [list, setList] = useState<any[]>([]); // 列表的数据
  const [loadStatus, setLoadStatus] = useState<LoadStatus>(LoadStatus.Loading);
  const [selections, setSelections] = useState<any[]>([]);

  // 单个 Agent 选中逻辑
  const handleItemCheck = item => {
    setSelections(original => {
      if (original.find(agent => agent.id === item.id)) {
        // 去除选中
        return original.filter(agent => agent.id !== item.id);
      } else {
        // 勾选
        return [...original, item];
      }
    });
  };

  // 请求获取agent列表
  const fetchAgents = useCallback(async ({ keyword = '' }: { keyword?: string } = {}) => {
    try {
      const { entries, pagination_marker_str } = await getAgentsByPost({
        pagination_marker_str: nextPaginationMarkerStrRef.current,
        size: pageSize,
        name: keyword,
        // 传1代表获取发布到自定义空间的智能体
        is_to_custom_space: 1,
      });

      // 请求用户头像
      const userIds = uniq(
        entries.map(item => item?.published_by).filter(userId => userId && !userAvatars[userId])
      ) as string[];

      if (userIds?.length) {
        addUserIds(userIds);
      }

      // 是否第一页
      const isFirstPage = !nextPaginationMarkerStrRef.current;
      nextPaginationMarkerStrRef.current = pagination_marker_str;

      if (isFirstPage) {
        setList(entries);
        setLoadStatus(entries.length > 0 ? LoadStatus.Normal : LoadStatus.Empty);
      } else {
        setList(original => [...original, ...entries]);
        setLoadStatus(LoadStatus.Normal);
      }
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
      setLoadStatus(LoadStatus.Failed);
    } finally {
      loadingMoreRef.current = false;
    }
  }, []);

  const debounceFetchData = useMemo(() => debounce(fetchAgents, 1000), [fetchAgents]);

  // 输入框value change
  const handleSearchValueChange = e => {
    // reset
    setLoadStatus(LoadStatus.Loading);
    setList([]);
    nextPaginationMarkerStrRef.current = '';
    loadingMoreRef.current = false;

    const value = e.target.value.trim();
    searchKeyRef.current = value;
    debounceFetchData({ keyword: value });
  };

  // 确认选择回调，可将选中数据传递给父组件
  const handleConfirm = () => {
    onConfirm(selections);
    onCancel(); // 也可根据需求决定是否关闭弹窗
  };

  const handleScroll = ({
    scrollTop, // 垂直滚动位置
  }) => {
    if (!gridRef.current) return;

    // 获取网格的总高度和可见高度
    const totalHeight = gridRef.current.props.rowHeight() * gridRef.current.props.rowCount;
    const clientHeight = gridRef.current.props.height;

    // 计算距离底部的距离
    const distanceFromBottom = totalHeight - scrollTop - clientHeight;

    // 判断是否接近底部（例如小于两行高度的像素）
    if (distanceFromBottom < 2 * rowHeight) {
      // 在这里可以触发加载更多数据的操作
      if (!loadingMoreRef.current && nextPaginationMarkerStrRef.current) {
        loadingMoreRef.current = true;
        setLoadStatus(LoadStatus.LoadingMore);
        debounceFetchData({ keyword: searchKeyRef.current });
      }
    }
  };

  const Cell = ({ columnIndex, rowIndex, style, originalRowCount }) => {
    // 渲染loading more行
    if (rowIndex === originalRowCount) {
      // 只在第一列显示，跨所有列
      if (columnIndex === 0) {
        return (
          <div
            style={{
              ...style,
              right: 0,
              width: 'unset',
              height: '20px',
              paddingRight: '20px',
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
            }}
          >
            <Spin size="small" />
          </div>
        );
      }

      return null;
    }

    const index = rowIndex * columnCount + columnIndex;

    // 边界检查：如果索引超出数据源长度，返回 null 不渲染任何内容
    if (index >= list.length) {
      return null;
    }

    // 渲染正常卡片内容
    const item = list[index];
    const time = '发布时间：' + formatTimeSlash(item?.published_at);

    return (
      <div
        style={{
          ...style,
          paddingRight: gap, // 右侧间隙
          paddingBottom: gap, // 底部间隙
        }}
      >
        <BaseCard
          checkable
          checked={selections.some(agent => agent.id === item.id)}
          bordered={true}
          hoverable={false}
          item={item}
          name={item?.name}
          profile={item?.profile}
          userName={item?.published_by_name}
          userAvatar={item?.published_by ? userAvatars[item?.published_by] : undefined}
          time={time}
          onCheckedChange={() => handleItemCheck(item)}
        />
      </div>
    );
  };

  const renderContent = () => {
    switch (loadStatus) {
      case LoadStatus.Loading:
        return (
          <div className={styles['loading']}>
            <Spin />
          </div>
        );

      case LoadStatus.Empty:
        return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />;

      case LoadStatus.Failed:
        return (
          <Empty
            image={
              <div>
                <LoadFailedIcon style={{ fontSize: '64px' }} />
                <div className={styles['failed-text']}>{intl.get('dataAgent.loadFailed')}</div>
              </div>
            }
            description=""
          />
        );

      default:
        //  Agent 列表区域
        return (
          <div style={{ height: '438px', width: 'calc(100% + 24px)' }}>
            <AutoSizer>
              {({ width, height }) => {
                const originalRowCount = Math.ceil(list.length / columnCount);
                // 添加一个额外的加载行
                const rowCount = loadStatus === LoadStatus.LoadingMore ? originalRowCount + 1 : originalRowCount;

                return (
                  <VariableSizeGrid
                    ref={gridRef}
                    overscanRowCount={3}
                    columnCount={columnCount}
                    rowCount={rowCount}
                    columnWidth={() => (width - 8) / columnCount}
                    rowHeight={index =>
                      index === originalRowCount && loadStatus === LoadStatus.LoadingMore
                        ? loadingMoreRowHeight
                        : rowHeight
                    }
                    width={width}
                    height={height}
                    className={styles['variable-size-grid']}
                    onScroll={handleScroll}
                  >
                    {props => Cell({ ...props, originalRowCount })}
                  </VariableSizeGrid>
                );
              }}
            </AutoSizer>
          </div>
        );
    }
  };

  useEffect(() => {
    debounceFetchData();
  }, []);

  useEffect(() => {
    if (gridRef.current) {
      // 重置网格缓存(否则 原先loading more 所在的那一行的 rowHeight 不会更新)
      gridRef.current.resetAfterIndices({
        rowIndex: Math.ceil(listCountRef.current / columnCount) - 1, // 重置受影响的行
        columnIndex: 0,
        shouldForceUpdate: true,
      });
    }
    listCountRef.current = list?.length;
  }, [list]);

  return (
    <Modal
      title={`${intl.get('dataAgent.selectAgent')} ${selections?.length ? intl.get('dataAgent.selectedCount', { count: selections?.length }) : ''}`}
      open={true}
      onCancel={onCancel}
      footer={
        <div className="dip-flex-space-between">
          <div
            className={classNames('dip-text-blue-link', {
              [styles['disabled']]: !selections?.length,
            })}
            onClick={() => setSelections([])}
          >
            {intl.get('dataAgent.deselect')}
          </div>
          <div className="dip-flex dip-gap-8">
            <Button
              key="confirm"
              type="primary"
              disabled={!selections?.length}
              onClick={handleConfirm}
              className={styles['btn']}
            >
              {intl.get('dataAgent.ok')}
            </Button>
            <Button key="cancel" onClick={onCancel} className={styles['btn']}>
              {intl.get('dataAgent.cancel')}
            </Button>
          </div>
        </div>
      }
      width={830}
      maskClosable={false}
      centered
    >
      {/* 搜索区域 */}
      <div className={styles['search']}>
        <Input
          placeholder={intl.get('dataAgent.searchAgent')}
          onChange={handleSearchValueChange}
          prefix={<SearchOutlined className="dip-opacity-75" />}
          allowClear
        />
      </div>
      {renderContent()}
    </Modal>
  );
};

export default AgentSelectionModal;
