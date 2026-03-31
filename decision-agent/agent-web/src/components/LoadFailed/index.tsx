import React from 'react';
import { Empty } from 'antd';
import intl from 'react-intl-universal';
import loadFailed from '@/assets/images/load-failed.png';

interface LoadFailedProps {
  className?: string;
  style?: React.CSSProperties;
  onRetry?: () => void;
}

const LoadFailed = ({ className, style, onRetry }: LoadFailedProps) => {
  return (
    <div className={className} style={style}>
      <Empty image={loadFailed} description={intl.get('dataAgent.loadFailed')} />
      {onRetry ? (
        <span className="dip-text-blue-link dip-mt-6" onClick={onRetry}>
          {intl.get('dataAgent.retry')}
        </span>
      ) : null}
    </div>
  );
};

export default LoadFailed;
