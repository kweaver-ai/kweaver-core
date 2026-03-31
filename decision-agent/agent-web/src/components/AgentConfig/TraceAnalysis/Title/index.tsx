import styles from './index.module.less';
import classNames from 'classnames';
import React from 'react';

export type TraceAnalysisTitleProps = {
  title: React.ReactNode;
  description?: React.ReactNode;
  className?: string;
  extra?: React.ReactNode;
  color?: string;
};
const TraceAnalysisTitle = (props: TraceAnalysisTitleProps) => {
  const { title, description, extra, className, color = '#126ee3' } = props;
  return (
    <div className={classNames('dip-flex-space-between dip-w-100', className)}>
      <div className="dip-flex-item-full-width">
        <div className={styles.title}>
          <div className={styles.line} style={{ backgroundColor: color }} />
          <span className="dip-flex-item-full-width dip-ml-8 dip-font-16">{title}</span>
        </div>
        {description && <div className="dip-mt-8 dip-font-12 dip-text-color-65">{description}</div>}
      </div>
      {extra}
    </div>
  );
};

export default TraceAnalysisTitle;
