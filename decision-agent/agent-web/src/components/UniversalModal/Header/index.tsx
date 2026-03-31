import React from 'react';
import classnames from 'classnames';
import Format from '@/components/Format';

export interface HeaderModalProps {
  children?: React.ReactNode;
  title: string | React.ReactNode; // 标题
  className?: string;
}

const Header = (props: HeaderModalProps) => {
  const { title, className } = props;

  return (
    <div className={classnames('ad-universal-modal-header dip-flex-align-center dip-pl-24', className)}>
      {title ? (
        typeof title !== 'string' ? (
          <div className="ad-format-text-no-height-3 ad-format-strong-6 dip-c-header">{title}</div>
        ) : (
          <Format.Title level={3} className="ad-format-text-no-height-3">
            {title}
          </Format.Title>
        )
      ) : null}
    </div>
  );
};

export default (props: any) => {
  const { visible = true, ...other } = props;
  if (!visible) return null;
  return <Header {...other} />;
};
