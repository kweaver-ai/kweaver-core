import React from 'react';
import { LoadingOutlined } from '@ant-design/icons';
import classNames from 'classnames';
import { useMicroWidgetProps } from '@/hooks';
import './style.less';

export interface LoadingMaskProps {
  classname?: string;
  style?: React.CSSProperties;
  loading: boolean;
  text?: string;
}

/**
 * loading状态，有一层遮罩防止点击, 外层需指定position：relative, 全覆盖
 */
const LoadingMask = (props: LoadingMaskProps) => {
  const { classname, style, loading, text } = props;
  const microWidgetProps = useMicroWidgetProps();
  const theme = microWidgetProps?.config?.getTheme?.normal;

  return loading ? (
    <div className={classNames('c-loading-mask', classname)} style={style}>
      <span className="c-l-icon dip-column-center">
        <LoadingOutlined style={{ fontSize: 24, color: theme }} />
        {text && (
          <span style={{ fontSize: 14 }} className="dip-c-subtext dip-mt-8">
            {text}
          </span>
        )}
      </span>
    </div>
  ) : null;
};

export default LoadingMask;
