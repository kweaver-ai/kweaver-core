/* eslint-disable wrap-iife */
import { type ReactNode, useCallback, useEffect, useMemo, useReducer, useRef } from 'react';
import ResizableHeader from './ResizableHeader';
import { useDebounceFn } from './utils/useDebounceFn';
import { depthFirstSearch, isEmpty } from './utils';
import { useSafeState } from './utils/useSafeState';
import { useLocalColumns } from './utils/useLocalColumns';
import { GETKEY } from './utils/useGetDataIndexColumns';
import { useMemoizedFn } from './utils/useMemoizedFn';
import { useLatest } from './utils/useLatest';
import { useUpdateThrottleEffect } from './utils/useUpdateThrottleEffect';
import { useIsomorphicLayoutEffect } from './utils/useIsomorphicLayoutEffect';
import { validateColumnsFlex } from './utils/validateOptions';

export interface ColumnsState {
  width: number;
}

export interface ColumnsStateType {
  /**
   * 持久化的类型，支持 localStorage 和 sessionStorage
   *
   * @param localStorage 设置在关闭浏览器后也是存在的
   * @param sessionStorage 关闭浏览器后会丢失
   */
  persistenceType?: 'localStorage' | 'sessionStorage';
  /** 持久化的key，用于存储到 storage 中 */
  persistenceKey?: string;
}

export interface OptionsType<ColumnType extends ColumnOriginType<ColumnType> = Record<string, any>> {
  columns: ColumnType[] | undefined;
  /** @description 最后一列不能拖动，设置最后一列的最小展示宽度，默认120 */
  defaultWidth?: number;
  /** @description 拖动最小宽度 默认0 */
  minConstraints?: number;
  /** @description 拖动最大宽度 默认无穷 */
  maxConstraints?: number;
  /** @description 是否缓存宽度 */
  cache?: boolean;
  /** @description 列状态的配置，可以用来操作列拖拽宽度 */
  columnsState?: ColumnsStateType;
  /** @description 开始拖拽时触发 */
  onResizeStart?: (col: ColumnType & { resizableColumns: ColumnType[] }) => void;
  /** @description 结束拖拽时触发 */
  onResizeEnd?: (col: ColumnType & { resizableColumns: ColumnType[] }) => void;
}

type WidthType = number | string;

export interface ColumnOriginType<T = any> {
  width?: WidthType;
  dataIndex?: string | number;
  key?: string | number;
  title?: ReactNode | string;
  children?: T[];
  resizable?: boolean; // internal
  ellipsis?: any;
  hideInTable?: boolean;
  maxWidth?: number;
}

interface CacheType {
  width?: WidthType;
  index: number;
}

const WIDTH = 120;

function useAntdResizableHeader<ColumnType extends ColumnOriginType<ColumnType>>(props: OptionsType<ColumnType>) {
  const {
    columns: columnsProp,
    defaultWidth = WIDTH,
    minConstraints = WIDTH / 2,
    maxConstraints = Number.POSITIVE_INFINITY,
    cache = true,
    columnsState,
    onResizeStart: onResizeStartProp,
    onResizeEnd: onResizeEndProp
  } = props;

  validateColumnsFlex(columnsProp);

  // add column width cache to avoid column's width reset after render
  const widthCache = useRef<Map<string | number, CacheType>>(new Map());

  // keep all default vlaue (e.g. defaultFilterValue)
  const [resizableColumns, setResizableColumns] = useSafeState<ColumnType[]>(columnsProp || []);

  const latestColumns = useLatest(resizableColumns);

  const { localColumns: columns, resetLocalColumns } = useLocalColumns({
    columnsState,
    columns: columnsProp,
    resizableColumns
  });

  const [tableWidth, setTableWidth] = useSafeState<number>();

  const [triggerRender, forceRender] = useReducer(s => s + 1, 0);

  let kvMap: Map<string | number, CacheType>;

  const resetColumns = useMemoizedFn(() => {
    kvMap = new Map();
    widthCache.current = kvMap;
    resetLocalColumns();
  });

  const onMount = useCallback(
    (id?: string | number) => (width?: number) => {
      if (width) {
        setResizableColumns(t => {
          const nextColumns = depthFirstSearch(t, col => col[GETKEY] === id && !!col.width, width);

          // console.log('nextColumns---', nextColumns);

          kvMap = kvMap || new Map<string | number, CacheType>();

          function dig(cols: ColumnType[]) {
            cols.forEach((col, i) => {
              const key = col[GETKEY];
              kvMap.set(key ?? '', { width: col?.width, index: i });
              if (col?.children) {
                dig(col.children);
              }
            });
          }

          dig(nextColumns);
          widthCache.current = kvMap;
          return nextColumns;
        });
      }
    },
    []
  );

  const onResize = useMemo(() => onMount, [onMount]);

  const onResizeStart = (col: ColumnType) => (width: number) => {
    onResizeStartProp?.({
      ...col,
      width,
      resizableColumns: latestColumns.current
    });
  };

  const onResizeEnd = (col: ColumnType) => (width: number) => {
    onResizeEndProp?.({
      ...col,
      width,
      resizableColumns: latestColumns.current
    });
  };

  const getColumns = useMemoizedFn((list: ColumnType[]) => {
    const trulyColumns = list?.filter(item => !isEmpty(item));

    const c = trulyColumns.map(col => {
      return {
        ...col,
        children: col?.children?.length ? getColumns(col.children) : undefined,
        onHeaderCell: (column: ColumnType) => {
          if (column.key === 'action') {
            // console.log('source column', column);
            return {
              'data-index': column.dataIndex,
              title: typeof col?.title === 'string' ? col?.title : '',
              width: cache ? widthCache.current?.get(column[GETKEY] ?? '')?.width || column?.width : column?.width,
              resizable: column.resizable,
              onMount: onMount(column?.[GETKEY]),
              onResize: onResize(column?.[GETKEY]),
              onResizeStart: onResizeStart(column),
              onResizeEnd: onResizeEnd(column),
              minWidth: 50,
              maxWidth: maxConstraints,
              triggerRender
            };
          }
          return {
            'data-index': column.dataIndex,
            title: typeof col?.title === 'string' ? col?.title : '',
            width: cache ? widthCache.current?.get(column[GETKEY] ?? '')?.width || column?.width : column?.width,
            resizable: column.resizable,
            onMount: onMount(column?.[GETKEY]),
            onResize: onResize(column?.[GETKEY]),
            onResizeStart: onResizeStart(column),
            onResizeEnd: onResizeEnd(column),
            minWidth: minConstraints,
            maxWidth: column?.maxWidth || maxConstraints,
            triggerRender
          };
        },
        width: cache ? widthCache.current?.get(col[GETKEY] ?? '')?.width || col?.width : col?.width,
        ellipsis: typeof col.ellipsis !== 'undefined' ? col.ellipsis : true,
        [GETKEY]: col[GETKEY] || col.key
      };
    }) as ColumnType[];

    return c;
  });

  useIsomorphicLayoutEffect(() => {
    if (columns) {
      const c = getColumns(columns);
      setResizableColumns(c);
    }
  }, [columns]);

  useUpdateThrottleEffect(
    () => {
      const t = getColumns(resizableColumns);
      setResizableColumns(t);
    },
    [triggerRender],
    {
      wait: 500
    }
  );

  useIsomorphicLayoutEffect(() => {
    let width = 0;
    (function loop(cls: ColumnType[]) {
      for (let i = 0; i < cls.length; i++) {
        if (cls[i].children) {
          loop(cls[i].children as ColumnType[]);
        } else if (!cls[i].hideInTable) {
          width += Number(cls[i].width) || defaultWidth;
        }
      }
    })(resizableColumns);
    setTableWidth(width);
  }, [defaultWidth, resizableColumns]);

  const { run: debounceRender } = useDebounceFn(forceRender);

  useEffect(() => {
    window.addEventListener('resize', debounceRender);
    return () => {
      window.removeEventListener('resize', debounceRender);
    };
  }, [debounceRender]);

  const components = useMemo(() => {
    return {
      header: {
        cell: ResizableHeader
      }
      // body: {
      //   cell: ResizableBody
      // }
    };
  }, []);

  return {
    resizableColumns: resizableColumns as any[],
    components,
    tableWidth,
    resetColumns
  };
}

export { useAntdResizableHeader };
