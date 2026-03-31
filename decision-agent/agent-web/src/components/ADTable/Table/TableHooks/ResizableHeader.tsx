/* eslint-disable no-empty-pattern */
import React, { type FC, type ThHTMLAttributes, memo, useEffect, useRef } from 'react';
import { Resizable, type ResizeCallbackData } from 'react-resizable';
import classnames from 'classnames';
import { useSafeState } from './utils/useSafeState';
import { type ColumnOriginType } from './useAntdResizableHeader';

import './index.css';

type ComponentProp = {
  onResize: (width: number) => void;
  onMount: (width: number) => void;
  onResizeStart?: (width: number) => void;
  onResizeEnd?: (width: number) => void;
  triggerRender: number;
  width: number;
  minWidth: number;
  maxWidth: number;
} & ColumnOriginType<any> &
  ThHTMLAttributes<HTMLTableCellElement>;

const ResizableHeader: FC<ComponentProp> = props => {
  const {
    width,
    minWidth,
    maxWidth,
    resizable,
    hideInTable,
    onResize,
    onResizeStart,
    onResizeEnd,
    onMount,
    triggerRender,
    className,
    style,
    onClick,
    children,
    rowSpan,
    colSpan,
    title,
    scope,
    ...rest
  } = props;

  const thRef = useRef<HTMLTableCellElement>(null);

  const [resizeWidth, setResizeWidth] = useSafeState<number>(0);

  useEffect(() => {
    if (width) {
      setResizeWidth(width);
      onMount?.(width);
    }
  }, [triggerRender]);

  useEffect(() => {
    if (width) {
      setResizeWidth(width);
    }
  }, [setResizeWidth, width]);

  if (hideInTable) {
    return null;
  }

  if (!width || Number.isNaN(Number(width)) || resizable === false) {
    return (
      <th
        {...rest}
        data-arh-disable="true"
        style={{ ...style }}
        className={classnames('ad-table-header', className)}
        onClick={onClick}
        rowSpan={rowSpan}
        colSpan={colSpan}
      >
        <span title={title}>{children}</span>
      </th>
    );
  }

  const setBodyStyle = (active: boolean) => {
    document.body.style.userSelect = active ? 'none' : '';
    document.body.style.pointerEvents = active ? 'none' : '';
    document.documentElement.style.cursor = active ? 'col-resize' : '';
  };

  const onStart = ({}, data: ResizeCallbackData) => {
    setResizeWidth(data.size.width);
    setBodyStyle(true);
    onResizeStart?.(data.size.width);
  };

  const onSelfResize = ({}, data: ResizeCallbackData) => {
    setResizeWidth(data.size.width);
  };

  const onStop = () => {
    if (resizeWidth <= 0) return;
    onResize(resizeWidth);
    setBodyStyle(false);
    onResizeEnd?.(resizeWidth);
  };

  return (
    <Resizable
      className="resizable-box"
      width={resizeWidth}
      minConstraints={[minWidth, 0]}
      maxConstraints={[maxWidth, 0]}
      height={0}
      handle={
        <div
          className="resizable-handler"
          onClick={e => {
            e.stopPropagation();
          }}
        >
          <div className="resizable-line" />
        </div>
      }
      draggableOpts={{ enableUserSelectHack: false }}
      onResizeStart={onStart}
      onResize={onSelfResize}
      onResizeStop={onStop}
    >
      <th
        {...rest}
        data-arh-disable="true"
        style={{ ...style }}
        className={classnames('ad-table-header', className)}
        onClick={onClick}
        rowSpan={rowSpan}
        colSpan={colSpan}
      >
        <span title={title}>{children}</span>
      </th>
    </Resizable>
  );
};

export default memo(ResizableHeader);
