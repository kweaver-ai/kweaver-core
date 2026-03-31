import React, { useEffect, useMemo, useRef, useState } from 'react';
import _ from 'lodash';
import classnames from 'classnames';
import intl from 'react-intl-universal';
import { Dropdown, Table } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';
import type { ColumnType } from 'antd/lib/table/interface';
import { nanoid } from 'nanoid';

import { useDeepCompareEffect } from '@/hooks';
import AdResizeObserver from '@/components/AdResizeObserver/AdResizeObserver';

import type { ITableProps } from '../types';
import useATRH from './TableHooks';

import EmptyIcon from '@/assets/icons/empty.svg';
import NoResultIcon from '@/assets/icons/no-result.svg';
import './style.less';

const ITable: React.FC<ITableProps> = props => {
  const {
    loading,
    width = 'auto',
    autoScrollY = false,
    contextMenu = {
      headerContextMenu: <></>,
      bodyContextMenu: <></>,
    },
    title,
    columns = [],
    dataSource: data,
    onRow,
    onHeaderRow,
    locale,
    emptySvg = null,
    emptyText = '',
    className,
    scroll = { x: 'max-content' },
    pagination,
    lastColWidth = 120,
    persistenceID = '',
    rowKey = 'key',
    filterToolsOptions,
    searchValue,
    onScrollCapture,
    activeRowKey,
    rowClassName,
    ...resetProps
  } = props;
  const EmptySvg = emptySvg;
  const { headerContextMenu, bodyContextMenu } = contextMenu;
  const TableRef = useRef<any>(null);
  const [menuPos, setMenuPos] = useState<string>('BODY_MENU');
  const [tableSize, setTableSize] = useState(0);
  const [showShadow, setShowShadow] = useState(false);
  const [colsID, setColsID] = useState(persistenceID);
  const [tableScrollY, setTableScrollY] = useState<number>(0);
  const tableWrapperRef = useRef<HTMLDivElement | null>(null);
  const domId = useRef<string>(nanoid());
  const scrollIntoViewAble = useRef(false);

  const renderCell = (value: any) => {
    return value || '--';
  };
  const handleColumn = (cols: ColumnType<any>[]) => {
    // const perColWidth = window.innerWidth * Number((1 / cols.length).toFixed(2));
    delete cols[cols?.length - 1]?.width;
    const withWidthCols = cols.filter(col => typeof col.width === 'number' || typeof col.width === 'string');
    if (withWidthCols.length === cols.length) return;
    const hasWidth = withWidthCols.reduce((pre, cur) => {
      return pre + Number(cur.width);
    }, 0);
    const perColWidth = (tableSize - hasWidth) * Number((1 / (cols.length - withWidthCols.length)).toFixed(2));

    const newColumns = cols.map((col: any, _index: any) => {
      if (col?.children && !_.isEmpty(col?.children)) {
        const perSubColWidth = Math.floor(perColWidth / col.children.length);
        col.children = _.map(col?.children, item => {
          return { ...item, width: perSubColWidth };
        });
      }

      if (_index === cols.length - 1) {
        if (col?.children && !_.isEmpty(col?.children)) {
          const maxWidth = 100;
          col.children[col.children.length - 1] = { ...col?.children[col?.children.length - 1], width: 0, maxWidth };
        }

        return { ...col, width: 0 };
      }

      return { ...col, width: col.width || perColWidth, render: col.render || renderCell };
    });
    return newColumns;
  };

  const newColumns = useMemo(() => handleColumn(columns), [columns, window, tableSize]);

  const {
    components,
    resizableColumns,
    tableWidth = 0,
    resetColumns,
  } = useATRH({
    columns: useMemo(() => {
      return newColumns as any;
    }, [columns, newColumns]),
    minConstraints: 60,
    defaultWidth: resetProps.rowSelection
      ? lastColWidth + Number(resetProps.rowSelection.columnWidth ?? 32)
      : lastColWidth,
    columnsState: !colsID
      ? undefined
      : {
          persistenceKey: `ADTableCols-${colsID}`,
          persistenceType: 'sessionStorage',
        },
  });

  // 持久化表单列宽
  useEffect(() => {
    if (!colsID) {
      // 去除前一次sessionStorage存的ID
      const length = sessionStorage.length;
      if (length > 0) {
        for (let i = 0; i < length; i++) {
          const key = sessionStorage.key(i);
          if (key?.includes('ADTableCols')) {
            sessionStorage.removeItem(key);
          }
        }
      }
      setColsID(nanoid().replace(/-/g, '_'));
    }
    return () => {};
  }, [TableRef, columns, persistenceID]);

  useEffect(() => {
    setTableSize(TableRef.current.clientWidth);
    resetColumns();
    // console.log('width', window.innerWidth * 0.25, 'height', window.innerHeight);
    return () => {
      resetColumns();
    };
  }, [TableRef, columns, persistenceID]);

  useEffect(() => {
    if (tableWidth > tableSize) {
      setShowShadow(true);
    } else {
      setShowShadow(false);
    }
  }, [tableWidth, tableSize, persistenceID]);

  // 刷新浏览器清楚拉动列宽
  useEffect(() => {
    window.onbeforeunload = function () {
      setTableSize(TableRef.current.clientWidth);
      resetColumns();
      // 去除前一次sessionStorage存的ID
      const length = sessionStorage.length;
      if (length > 0) {
        for (let i = 0; i < length; i++) {
          const key = sessionStorage.key(i);
          if (key?.includes('ADTableCols')) {
            sessionStorage.removeItem(key);
          }
        }
      }
    };
    return () => {
      window.onbeforeunload = null;
    };
  }, []);

  useDeepCompareEffect(() => {
    handleActiveRowScrollIntoView();
  }, [data]);

  const paginationConfig = useMemo(() => {
    if (pagination) {
      return {
        showTotal: (total: number) => intl.get('knowledge.total', { total }),
        showSizeChanger: false,
        hideOnSinglePage: true,
        ...pagination,
      };
    }
    return false;
  }, [pagination]);

  const prefixCls = 'ad-table';

  const isAutoCalculateScrollY = autoScrollY && !scroll.y && !!data && data.length > 0; // 是否自动计算scrollY的值

  useDeepCompareEffect(() => {
    if (isAutoCalculateScrollY && paginationConfig) {
      if (paginationConfig.total! > paginationConfig.pageSize!) {
        onResize();
      }
    }
  }, [isAutoCalculateScrollY, data, paginationConfig]);

  const handleActiveRowScrollIntoView = () => {
    const selectedCardDOM = document.querySelector(`.${prefixCls}-row-active`);
    if (selectedCardDOM && !scrollIntoViewAble.current) {
      selectedCardDOM.scrollIntoView();
      scrollIntoViewAble.current = true;
    }
  };

  /** 获取缺醒图 */
  const getLocale = () => {
    // loading的时候  不显示缺省图
    if (loading || locale === null) {
      return { emptyText: ' ' };
    }
    // dataSource为空，但是存在搜索或者过滤条件，显示空数据
    const filterValue = filterToolsOptions?.find(item => !!item.value && item.value !== 'all');

    let showEmptyData = false;
    if (searchValue || filterValue) {
      showEmptyData = true;
    }
    if (showEmptyData) {
      return {
        emptyText: (
          <div className={`${prefixCls}-nodata-box`}>
            <NoResultIcon />
            <div className={`${prefixCls}-nodata-text`}>{intl.get('global.noResult')}</div>
          </div>
        ),
      };
    }
    return (
      locale || {
        emptyText: (
          <div className={`${prefixCls}-nodata-box`}>
            {EmptySvg ? <EmptySvg /> : <EmptyIcon />}
            <div className={`${prefixCls}-nodata-text`}>{emptyText || intl.get('global.noContent')}</div>
          </div>
        ),
      }
    );
  };

  const targetColumns = useMemo(() => {
    if (resizableColumns && resizableColumns.length > 0) {
      resizableColumns[resizableColumns.length - 1].render =
        resizableColumns[resizableColumns.length - 1].render || renderCell;
      return resizableColumns;
    }
    return [];
  }, [resizableColumns]);

  const getLoading = () => {
    if ((isAutoCalculateScrollY && tableScrollY === 0) || loading) {
      // 说明此时还没有计算出scrollY的值
      return {
        indicator: <LoadingOutlined className="icon" style={{ fontSize: 24, top: '200px' }} spin />,
      };
    }
    return false;
  };

  const renderContent = () => {
    return (
      <div
        style={{ width }}
        onContextMenu={e => {
          e.preventDefault();
          e.stopPropagation();
        }}
        className="ad-table-container dip-flex-item-full-height"
        id={domId.current}
      >
        <Dropdown
          placement="bottomLeft"
          overlay={() => {
            if (menuPos === 'BODY_MENU') {
              return bodyContextMenu;
            }
            return headerContextMenu;
          }}
          trigger={['contextMenu']}
        >
          <div ref={tableWrapperRef} onScrollCapture={onScrollCapture}>
            <Table
              title={title}
              className={classnames(prefixCls, { 'no-shadow': !showShadow }, className)}
              loading={getLoading()}
              ref={TableRef}
              columns={targetColumns}
              components={components}
              dataSource={data}
              scroll={{ x: tableWidth, y: isAutoCalculateScrollY ? tableScrollY : scroll.y }}
              rowKey={rowKey}
              locale={getLocale()}
              onRow={(record, index) => {
                const RowContextMenuObj = onRow?.(record, index);
                const RowcontextMenuCallBack = RowContextMenuObj?.onContextMenu || (() => {});
                return {
                  ...RowContextMenuObj,
                  onContextMenu: e => {
                    e.preventDefault();
                    setMenuPos('BODY_MENU');
                    RowcontextMenuCallBack(e);
                  },
                };
              }}
              onHeaderRow={(columns, index) => {
                const headerRowContextMenuObj = onHeaderRow?.(columns, index);
                const headerRowContextMenuCallBack = headerRowContextMenuObj?.onContextMenu || (() => {});

                return {
                  ...headerRowContextMenuObj,
                  onContextMenu: e => {
                    e.preventDefault();
                    setMenuPos('HEADER_MENU');
                    headerRowContextMenuCallBack(e);
                  },
                };
              }}
              pagination={paginationConfig}
              rowClassName={(record, index, indent) => {
                let key = record.key;
                if (rowKey) {
                  if (typeof rowKey === 'string') {
                    key = record[rowKey];
                  } else {
                    key = rowKey(record);
                  }
                }
                const rowCls =
                  typeof rowClassName === 'function' ? rowClassName?.(record, index, indent) : rowClassName;
                return classnames(rowCls, {
                  [`${prefixCls}-row-active`]: activeRowKey !== undefined && key !== undefined && activeRowKey === key,
                });
              }}
              {...resetProps}
            />
          </div>
        </Dropdown>
      </div>
    );
  };

  /**
   * 获取DOM元素的高度（包括margin的高度）
   * @param domElement
   */
  const getDomHeight = (domElement: Element) => {
    const computedStyle = window.getComputedStyle(domElement);
    const elementHeight =
      parseFloat(computedStyle.height) + parseFloat(computedStyle.marginTop) + parseFloat(computedStyle.marginBottom);
    return elementHeight || 0;
  };
  const onResize = _.debounce(() => {
    let scrollY = 0;
    const domRefDom = document.getElementById(domId.current);
    const tableTitleDom = tableWrapperRef.current?.querySelector('.agent-web-table-title');
    const tableHeaderDom = tableWrapperRef.current?.querySelector('.agent-web-table-thead');
    const tablePaginationDom = tableWrapperRef.current?.querySelector('.agent-web-pagination'); // 52px
    if (domRefDom) {
      scrollY = getDomHeight(domRefDom);
    }
    if (tableTitleDom) {
      scrollY -= getDomHeight(tableTitleDom);
    }
    if (tableHeaderDom) {
      scrollY -= getDomHeight(tableHeaderDom);
    }
    if (tablePaginationDom) {
      scrollY -= getDomHeight(tablePaginationDom);
    }
    setTableScrollY(scrollY);
  }, 300);

  return isAutoCalculateScrollY ? (
    <AdResizeObserver onResize={onResize}>{renderContent()}</AdResizeObserver>
  ) : (
    renderContent()
  );
};

export default ITable;
